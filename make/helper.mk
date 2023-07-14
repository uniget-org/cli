.PHONY:
$(addprefix helper--,$(ALL_TOOLS_RAW)):helper--%: \
		$(HELPER)/var/lib/uniget/manifests/%.json

$(HELPER)/var/lib/uniget/manifests/%.json:
	@if ! type uniget >/dev/null 2>&1; then \
		echo "Please install uniget"; \
		exit 1; \
	fi
	@set -o errexit; \
	mkdir -p $(HELPER)/var/cache $(HELPER)/var/lib $(HELPER)/usr/local; \
	uniget --prefix=$$PWD/$(HELPER) update; \
	uniget --prefix=$$PWD/$(HELPER) install $*