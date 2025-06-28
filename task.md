implement a proper token counting system.

count the tokens in the input. process the input with every model that has a greater context window than the input token count.

cleanly log every model's "attempt" at processing the input, and for every model with a context window too small to process the input log an info message saying something like "skipping this model because the input is too large for its context window"

use our modern clean logging system to log a bit of info about our token counting too, for visibility, and obviously update our structured logging to thinktank.log and the audit file
