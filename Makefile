.PHONY: install-hook # build

# Install hook is not currently meaningfully used since I haven't figured out how to bundle the version
# with the binary.
install-hook:
	@cp pre-commit.sh .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Pre-commit hook installed."

# build:
# 	go build -o cir

# install: build
# 	go install github.com/worldsayshi/cir@latest
