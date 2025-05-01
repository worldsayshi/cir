
A LLM Chat client on the terminal built with `tview`.

Status: Alpha-ish

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

Type '?' to get key bindings.

# Run from this repo

Run:
```bash
go run .
```

Test:
```bash
go test ./...
```

# Backlog

## Better engine?

- Maybe use the [cview fork instead](https://codeberg.org/tslocum/cview)
    - it seems to have more maintainers and has merged more PR:s

## Styling of chat history

- Make chathistory nicer to read and navigate


## Stream and diff files

- We don't need to stream it with diff, just show the diff once the stream is done.
- Can even have a toggle for toggling between diff and normal view?

So:
- Whenever we're in a code section, stop rendering in the history view and stream into some kind of buffer?
- The buffer can be opened when selecting it somehow.
- Should the chat history be a list of messages?
- Pressing enter on an element shows a list of code blocks that can be looked at, and diffed if there's a corresponding file.


## Simpler session selector

Mostly feature complete but some more polish can be applied.

- [/] Find all files that can be opened as sessions and select one
    - Store `kind: WorkingSession` for making it easier to find them
    - Search upwards in file tree, see https://claude.ai/chat/0946b93d-a0e6-49cf-a0ba-3eece3f30b5e
- [X] Use key shortcut at first? Have a sane default to open, then use shortcut to open some other session.
- [ ] Load session on selection! <-- HERE!
- Key to create new session?

## TODOs for Beta

- [/] Allow code edits
    - Let's just start by parsing out files from answer.
    - Do I need edit mode?
        - shortcut key -> toggle edit mode
        - edit mode:
            - add system message -> explanation + syntax?
            - on response, split out edits from returned prompt
- [ ] Summary of keymaps, when pressing '?'
- [ ] More context info
    - [ ] Add the file names sent to the printed chat message
- [ ] Bug: Getting `Error: <nil>` in log
- [ ] Cleanup: Get rid of frivolous panics
- [/] Refactor application.go so that the control flow is more DAG-like, now it's spaghet
    - Take inspo from this conversation maybe: https://claude.ai/chat/9efbb9f6-4bbc-48e7-ac35-f825dbdae7d9

## Session selector

- [ ] Session selector v1: Multiple session switcher and multiple sessions in .cir folder in ./cir/sessions/
    - [ ] Allow migrating from a single session to multiple
    - [ ] Select the default session
    - [] Keep track of cwd for the session
- [ ] Session selector v2:
    - [ ] Keep track of all sessions that have been opened by adding a session reference in ~/.cir/sessions.yaml¨
    - [ ] Deal with moved sessions somehow

## Nice to have's

- [ ] Integrate Copilot API <-- Good exercise!!
    - Reference 1: https://github.com/B00TK1D/copilot-api/blob/main/api.py
    - Reference 2: /rubberduck.vim/lua/copilot_request.lua
- [ ] Plugins like [k9s plugins](https://k9scli.io/topics/plugins/)?
- [ ] claude api support

# Done

## Alpha

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
