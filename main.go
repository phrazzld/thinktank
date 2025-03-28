// main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/phrazzld/architect/internal/fileutil"

	genai "github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const defaultOutputFile = "PLAN.md"
const defaultModel = "gemini-2.5-pro-exp-03-25"
const apiKeyEnvVar = "GEMINI_API_KEY"
const defaultFormat = "<{path}>\n```\n{content}\n```\n</{path}>\n\n"

// Default excludes inspired by common project types and handoff's defaults
const defaultExcludes = ".exe,.bin,.obj,.o,.a,.lib,.so,.dll,.dylib,.class,.jar,.pyc,.pyo,.pyd,.zip,.tar,.gz,.rar,.7z,.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.odt,.ods,.odp,.jpg,.jpeg,.png,.gif,.bmp,.tiff,.svg,.mp3,.wav,.ogg,.mp4,.avi,.mov,.wmv,.flv,.iso,.img,.dmg,.db,.sqlite,.log"
const defaultExcludeNames = ".git,.hg,.svn,node_modules,bower_components,vendor,target,dist,build,out,tmp,coverage,__pycache__,*.pyc,*.pyo,.DS_Store,~$*,desktop.ini,Thumbs.db,package-lock.json,yarn.lock,go.sum,go.work"


func main() {
	// --- Configuration & Flag Parsing ---
	task := flag.String("task", "", "Required: Description of the task or goal for the plan.")
	outputFile := flag.String("output", defaultOutputFile, "Output file path for the generated plan.")
	modelName := flag.String("model", defaultModel, "Gemini model to use for generation.")
	verbose := flag.Bool("verbose", false, "Enable verbose logging output.")
	include := flag.String("include", "", "Comma-separated list of file extensions to include (e.g., .go,.md)")
	exclude := flag.String("exclude", defaultExcludes, "Comma-separated list of file extensions to exclude.")
	excludeNames := flag.String("exclude-names", defaultExcludeNames, "Comma-separated list of file/dir names to exclude.")
	format := flag.String("format", defaultFormat, "Format string for each file. Use {path} and {content}.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s --task \"<your task description>\" [options] <path1> [path2...]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  <path1> [path2...]   One or more file or directory paths for project context.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  %s: Required. Your Google AI Gemini API key.\n", apiKeyEnvVar)
	}

	flag.Parse()

	// --- Input Validation ---
	if *task == "" {
		log.Println("Error: --task flag is required.")
		flag.Usage()
		os.Exit(1)
	}

	paths := flag.Args()
	if len(paths) == 0 {
		log.Println("Error: At least one file or directory path must be provided as an argument.")
		flag.Usage()
		os.Exit(1)
	}

	apiKey := os.Getenv(apiKeyEnvVar)
	if apiKey == "" {
		log.Printf("Error: %s environment variable not set.", apiKeyEnvVar)
		flag.Usage()
		os.Exit(1)
	}

	// --- Logger Setup ---
	logFlags := 0
	logPrefix := "[architect] "
	// Customize logger based on verbosity
	logger := log.New(os.Stderr, logPrefix, logFlags)
	if *verbose {
		logger.SetFlags(log.LstdFlags | log.Lmicroseconds) // More detail if verbose
		logger.SetPrefix("[architect-v] ")
	}


	// --- Gather Context using internal fileutil ---
	logger.Println("Gathering project context...")
	fileConfig := fileutil.NewConfig(*verbose, *include, *exclude, *excludeNames, *format, logger)

	projectContext, processedFilesCount, err := fileutil.GatherProjectContext(paths, fileConfig)
	if err != nil {
		// GatherProjectContext logs specific errors, maybe just exit?
		logger.Fatalf("Failed during project context gathering: %v", err) // Or handle more gracefully
	}

	if processedFilesCount == 0 {
		logger.Println("Warning: No files were processed for context. Check paths and filters.")
		// Optionally exit here, or let it proceed (Gemini might still work with just the task)
		// os.Exit(1)
	}

	// Log statistics if verbose
	if *verbose || processedFilesCount > 0 { // Always log stats if files were processed
		charCount, lineCount, tokenCount := fileutil.CalculateStatistics(projectContext)
		logger.Printf("Context gathered: %d files processed, %d lines, %d chars, ~%d tokens.", processedFilesCount, lineCount, charCount, tokenCount)
	}


	// --- Construct Gemini Prompt ---
	prompt := buildPrompt(*task, projectContext)
	if *verbose {
		logger.Printf("Prompt length: %d characters", len(prompt))
		logger.Printf("Sending task to Gemini: %s", *task)
	}


	// --- Initialize Gemini Client ---
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		logger.Fatalf("Error creating Gemini client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel(*modelName)
	model.SetMaxOutputTokens(8192) // Set a reasonable max output for plans
	model.SetTemperature(0.3)      // Slightly lower temperature for more deterministic plans
	model.SetTopP(0.9)             // Keep some creativity
	// Consider adding safety settings if needed


	// --- Call Gemini API ---
	logger.Printf("Generating plan using model %s...", *modelName)

	// Create content using the model directly
	resp, err := model.GenerateContent(ctx,
		genai.Text(prompt))

	// --- Process Response ---
    // Add more robust error checking for API call
    if err != nil {
        logger.Fatalf("Error generating content: %v", err)
    }

    if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
         // Check finish reason if available
        finishReason := ""
        if resp != nil && len(resp.Candidates) > 0 {
            finishReason = fmt.Sprintf(" (Finish Reason: %s)", resp.Candidates[0].FinishReason)
        }
        // Check safety ratings if available
        safetyInfo := ""
         if resp != nil && len(resp.Candidates) > 0 && resp.Candidates[0].SafetyRatings != nil {
            blocked := false
            for _, rating := range resp.Candidates[0].SafetyRatings {
                if rating.Blocked {
                    blocked = true
                    safetyInfo += fmt.Sprintf(" Blocked by Safety Category: %s;", rating.Category)
                }
            }
             if blocked {
                 safetyInfo = " Safety Blocking:" + safetyInfo
             }
        }
        logger.Fatalf("Error: Received invalid or empty response from Gemini.%s%s", finishReason, safetyInfo)
    }


	// Extract the text content
	generatedPlan := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			generatedPlan += string(textPart)
		} else {
            logger.Printf("Warning: Received non-text part in response: %T", part)
        }
	}

    if strings.TrimSpace(generatedPlan) == "" {
        logger.Fatalf("Error: Gemini returned an empty plan text.")
    }

	if *verbose {
		logger.Println("Plan received from Gemini.")
	}


	// --- Write Output File ---
	outputPath := *outputFile
	if !filepath.IsAbs(outputPath) {
		cwd, err := os.Getwd()
		if err != nil {
			logger.Fatalf("Error getting current working directory: %v", err)
		}
		outputPath = filepath.Join(cwd, outputPath)
	}

	logger.Printf("Writing plan to %s...", outputPath)
	err = os.WriteFile(outputPath, []byte(generatedPlan), 0644)
	if err != nil {
		logger.Fatalf("Error writing plan to file %s: %v", outputPath, err)
	}

	logger.Printf("Successfully generated plan and saved to %s", outputPath)
}

// buildPrompt constructs the prompt string for the Gemini API.
func buildPrompt(task string, context string) string {
	// Same prompt structure as before
	return fmt.Sprintf(`You are an expert software architect and senior engineer.
Your goal is to create a detailed, actionable technical plan in Markdown format.

**Task:**
%s

**Project Context:**
Below is the relevant code context from the project. Analyze it carefully to understand the current state.
%s

**Instructions:**
Based on the task and the provided context, generate a technical plan named PLAN.md. The plan should include the following sections:

1.  **Overview:** Briefly explain the goal of the plan and the changes involved.
2.  **Task Breakdown:** A detailed list of specific, sequential tasks required to implement the feature or fix.
    *   For each task, estimate the effort (e.g., S, M, L) or time.
    *   Mention the primary files/modules likely to be affected.
3.  **Implementation Details:** Provide specific guidance for the more complex tasks. Include:
    *   Key functions, classes, or components to modify or create.
    *   Data structures or API changes needed.
    *   Code snippets or pseudocode where helpful.
4.  **Potential Challenges & Considerations:** Identify possible risks, edge cases, dependencies, or areas needing further investigation.
5.  **Testing Strategy:** Outline how the changes should be tested (unit tests, integration tests, manual testing steps).
6.  **Open Questions:** List any ambiguities or points needing clarification before starting implementation.

Format the entire response as a single Markdown document suitable for direct use as `+"`PLAN.md`"+`. Do not include any introductory or concluding remarks outside the Markdown plan itself. Ensure the markdown is well-formatted.
`, task, context) // context already has the <context> tags from fileutil
}
