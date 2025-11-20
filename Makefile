up:
	docker compose down
	docker compose up --build -d
	docker image prune -f

ps:
	docker ps --filter "label=com.docker.compose.project=backend" --format "table {{.Names}}\t{{.Status}}"

log:
	git log --oneline -n 10