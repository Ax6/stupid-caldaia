APP_IMAGE=stupid-caldaia
CONTROLLER_IMAGE=stupid-caldaia-controller
WORKER_IMAGE=stupid-caldaia-worker

build-app:
	docker build -t $(APP_IMAGE) -f dockerfiles/app.Dockerfile app

build-controller:
	docker build -t $(CONTROLLER_IMAGE) -f dockerfiles/controller.Dockerfile .

build-worker:
	docker build -t $(WORKER_IMAGE) -f dockerfiles/worker.Dockerfile .

install-redis:
	curl -fsSL https://packages.redis.io/gpg | sudo gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg
	sudo chmod 644 /usr/share/keyrings/redis-archive-keyring.gpg
	echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/redis.list
	sudo apt-get update
	sudo apt-get install redis-stack-server