# BACKLOG

- transform into very simple interactive tui
	* steps
		* define task
		* select files / directories for additional context
		* select models / group to execute task
		* select output directory (or go with default)
		* confirm and run
- user should be able to save task prompts in their config
- user should be able to define an arbitrary number of steps in their task
	* ie task is 1) generate a plan file, 2) critique the plan file, 3) generate a second draft plan file
- user should be able to add a _synthesize_ step at the end
	* pick a model to send all of the model outputs to for synthesis
- better default config init
- program still hangs for a while after completing a run
