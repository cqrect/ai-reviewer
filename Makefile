release:
	@docker build --platform linux/amd64 -t cqrect/ai-reviewer:latest .
	@docker push cqrect/ai-reviewer:latest
.PHONY: release
