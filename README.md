
Status: WIP


# TODOs

- [X] Bug: refactor and fix messages handling so that messages are updated properly
- [/] adding context files using fzf?
    - [X] Define working set data in session data structure:
        - workingFiles with path and last submitted checksum
    - [X] Manage current working set in TUI - store in session file
    - [ ] Just add them to the message
        -
    - Add them but hide them from rendering
    - Add them to the session data, calculate and store checksum whenever they are added to the sent data, only send when checksum change
- [ ] prompt templates for context
- [ ] Make the history view scrollable
- [ ] Plugins like [k9s plugins](https://k9scli.io/topics/plugins/)?
- [ ] claude api support
- [ ] Bug: Getting `Error: <nil>` in log
- [ ] Get rid of panics