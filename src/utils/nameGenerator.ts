/**
 * Name generator module for thinktank
 * 
 * Provides functions to generate user-friendly run names using Google's Gemini-2.0-flash model
 * and fallback mechanisms when API generation fails.
 */
import { GoogleGenerativeAI, HarmCategory, HarmBlockThreshold } from "@google/generative-ai";
import { logger } from './logger';

// Removed Schema definition as we're using simple prompt-based approach

/**
 * Generates a fun, unique name for a run using Google's Gemini-2.0-flash model
 * Format is 'adjective-noun' (e.g., clever-meadow, swift-breeze)
 * 
 * @returns A fun name string if successful, null if generation fails
 */
export async function generateFunName(): Promise<string | null> {
  // Try to get Google API key
  const apiKey = process.env.GEMINI_API_KEY || process.env.GOOGLE_API_KEY;
  if (!apiKey) {
    logger.debug("GEMINI_API_KEY not found. Skipping fun name generation.");
    return null;
  }

  try {
    const genAI = new GoogleGenerativeAI(apiKey);
    const model = genAI.getGenerativeModel({
      model: "gemini-2.0-flash",
      generationConfig: {
        temperature: 0.7,
      },
      // Basic safety settings
      safetySettings: [
        { category: HarmCategory.HARM_CATEGORY_HARASSMENT, threshold: HarmBlockThreshold.BLOCK_MEDIUM_AND_ABOVE },
        { category: HarmCategory.HARM_CATEGORY_HATE_SPEECH, threshold: HarmBlockThreshold.BLOCK_MEDIUM_AND_ABOVE },
        { category: HarmCategory.HARM_CATEGORY_SEXUALLY_EXPLICIT, threshold: HarmBlockThreshold.BLOCK_MEDIUM_AND_ABOVE },
        { category: HarmCategory.HARM_CATEGORY_DANGEROUS_CONTENT, threshold: HarmBlockThreshold.BLOCK_MEDIUM_AND_ABOVE },
      ],
    });

    const prompt = "Generate a fun, unique, two-word name for a software project run. Format: 'adjective-noun', lowercase, hyphenated. Examples: 'clever-otter', 'sunny-vista'. Please provide only the name, nothing else.";
    
    // Request output
    let response;
    try {
      // Gemini models have different API behaviors - try safest approach first
      const result = await model.generateContent(prompt);
      response = result.response;
    } catch (error) {
      logger.debug("Error generating name with Gemini");
      return null;
    }

    let name = "";
    try {
      // Try to parse as JSON first (if structured output worked)
      const responseText = response.text();
      const parsedResponse = JSON.parse(responseText) as { name?: string };
      if (typeof parsedResponse.name === 'string') {
        name = parsedResponse.name;
      } else {
        // If name property is missing or not a string, use raw text
        name = responseText.trim();
      }
    } catch (parseError) {
      // If JSON parsing fails, use the raw text
      name = response.text().trim();
    }

    // Validate the name format
    if (name && typeof name === 'string' && /^[a-z]+-[a-z]+$/.test(name)) {
      return name;
    } else {
      logger.debug(`Received invalid name format from Gemini: ${name}`);
      // Try to extract a valid name from the text if possible
      const match = name.match(/[a-z]+-[a-z]+/);
      if (match && match[0]) {
        return match[0];
      }
      return null;
    }
  } catch (error) {
    logger.debug("Error generating fun name from Gemini");
    return null;
  }
}

/**
 * Generates a fallback name based on timestamp
 * Used when the Gemini API call fails
 * 
 * @returns A timestamp-based name (e.g., "run-20250101-123045")
 */
export function generateFallbackName(): string {
  const now = new Date();
  const timestamp = now.toISOString()
    .replace(/[-:T.Z]/g, '') // Remove separators
    .substring(0, 14); // YYYYMMDDHHmmss
  return `run-${timestamp.substring(0, 8)}-${timestamp.substring(8)}`; // Format as run-YYYYMMDD-HHmmss
}