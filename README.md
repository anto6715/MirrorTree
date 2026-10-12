# Mirror Tree

GoLang tool capable to mirror a given root dir and its contents to a target dir,
creating mock files and directories.

## Why Mock mirroring?

There are some scenario where only tree structure and filename matters, for example:

- check on file presence
- check on rolling rules triggered by filename
- archive rules based on filename
- ecc..

On all this scenario, having a Tree Mockup can help simulating
the target dir structure without actually creating files.
