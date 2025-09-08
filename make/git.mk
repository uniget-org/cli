BRANCHES := $(shell git ls-remote --heads 2>/dev/null | grep /renovate/ | cut -f2 | cut -d/ -f3-)

.PHONY:
branches:
	@echo $(BRANCHES)

.PHONY:
$(addprefix rebase--,$(BRANCHES)):rebase--%: ; $(info $(M) Rebasing branch $*...)
	@git switch $*
	@git reset --hard origin/$*
	@git rebase main
	@git push --force-with-lease
	@git switch main