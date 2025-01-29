
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

# Run from this repo

Run:
```bash
go run .
```

Test:
```bash
go test ./...
```

# TODOs for Beta

- [/] Refactor application.go so that the control flow is more DAG-like, now it's spaghet
    - Take inspo from this conversation maybe: https://claude.ai/chat/9efbb9f6-4bbc-48e7-ac35-f825dbdae7d9
- [ ] More context info
    - [ ] Add the file names sent to the printed chat message
- [ ] Allow code edits
- [ ] Bug: Getting `Error: <nil>` in log
- [ ] Cleanup: Get rid of frivolous panics

# Alpha TODO log

- [X] Bug: refactor and fix messages handling so that messages are updated properly
- [X] adding context files using fzf?
    - [X] Define working set data in session data structure:
        - workingFiles with path and last submitted checksum
    - [X] Manage current working set in TUI - store in session file
    - [X] Just add them to the message
    - [X] Add them but hide them from rendering
    - [X] Add them to the session data, calculate and store checksum whenever they are added to the sent data, only send when checksum change (sort of untested but seems to work I think??)
- [X] Add backwards compatibility for session yaml storage
- [X] Prompt templates for sending context
- [X] Make the history view scrollable
- [X] QOL: Also store the current wip chat message in the session (on exit?)

# Nice to have's

- [ ] Integrate Copilot API <-- Good exercise!!
    - Reference 1: https://github.com/B00TK1D/copilot-api/blob/main/api.py
    - Reference 2: /rubberduck.vim/lua/copilot_request.lua
- [ ] Plugins like [k9s plugins](https://k9scli.io/topics/plugins/)?
- [ ] claude api support