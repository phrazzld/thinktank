# BACKLOG

- simplify cascading config, fully centralize to user config
- decouple dependencies with interfaces. remove all http mocking and replace with proper dependency injections
- isolate side effects. separate i/o from logic.
- set up github actions
- program still hangs for a while after completing a run
- better default config init
- running w/o specifying a group should run the default group models -- not every fucking enabled model in the config lol
- user should be able to save task prompts in their config
- user should be able to define an arbitrary number of steps in their task
	* ie task is 1) generate a plan file, 2) critique the plan file, 3) generate a second draft plan file
- user should be able to add a _synthesize_ step at the end
	* pick a model to send all of the model outputs to for synthesis
- fix error: `Error from openai:o3-mini: (0 , errors_1.isProviderRateLimitError) is not a function`
- improve cli ui/ux
	* show progress indicator for each running model, spinners when in progress -> checkmarks or green circles when completed
	* fix "double i" problem (two i icons showing for a lot of the info messages)
- support ad-hoc task/prompt definitions (ie passing a string instead of a filepath)
- make it easy to write output to logfile
