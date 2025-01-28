
A LLM Chat client on the terminal built with `tview`.

Status: Pre-Alpha

# Requirements

- fzf-tmux
- `OPENAI_API_KEY` as env variable

# Install

```
go install github.com/worldsayshi/cir@main
```

Then just run `cir` to start it. It will store the session in a `~/.cir` folder.
For a separate session, use the `-session` flag like this: `cir -session my-session.yaml`.


# Key bindings

- Ctrl-o - Manage context
- Ctrl-s - Submit message
- (Shift-)Tab - Toggle focus between input and chat history

# TODOs for Alpha

- [X] Bug: refactor and fix messages handling so that messages are updated properly
- [/] adding context files using fzf?
    - [X] Define working set data in session data structure:
        - workingFiles with path and last submitted checksum
    - [X] Manage current working set in TUI - store in session file
    - [X] Just add them to the message
        -
    - [X] Add them but hide them from rendering
    - [ ] Add the file names sent to the printed chat message
    - [/] Add them to the session data, calculate and store checksum whenever they are added to the sent data, only send when checksum change
- [X] Prompt templates for sending context
- [X] Make the history view scrollable
- [X] QOL: Also store the current wip chat message in the session (on exit?)

# TODOs for Beta

- [/] Refactor application.go so that the control flow is more DAG-like, now it's spaghet
    - Take inspo from this conversation maybe: https://claude.ai/chat/9efbb9f6-4bbc-48e7-ac35-f825dbdae7d9
- [ ] Allow code edits
- [ ] Bug: Getting `Error: <nil>` in log
- [ ] Get rid of panics

# Nice to have's

- [ ] Integrate Copilot API <-- Good exercise!!
    - Reference 1: https://github.com/B00TK1D/copilot-api/blob/main/api.py
    - Reference 2: /home/perfr/workspace/rubberduck.vim/lua/copilot_request.lua
- [ ] Plugins like [k9s plugins](https://k9scli.io/topics/plugins/)?
- [ ] claude api support