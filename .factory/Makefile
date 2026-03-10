# Pipeline-critical targets used by TaintFactory.
# DO NOT modify or remove — these are called during automated pipeline setup.
# See CLAUDE.md "Pipeline-Critical Commands" for details.

.PHONY: migrate-up

migrate-up:
	@if command -v migrate >/dev/null 2>&1; then \
		set -a; \
		[ -f config/.env ] && . ./config/.env; \
		[ -f .env ] && . ./.env; \
		set +a; \
		migrate -database "$$DATABASE_URL" -path migrations/ up; \
	else echo "migrate not installed, skipping"; fi
