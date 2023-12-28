APP_IMAGE=stupid-caldaia
CONTROLLER_IMAGE=stupid-caldaia-controller
WORKER_IMAGE=stupid-caldaia-worker

build-app:
	docker build -t $(APP_IMAGE) -f dockerfiles/app.Dockerfile app

build-controller:
	docker build -t $(CONTROLLER_IMAGE) -f dockerfiles/controller.Dockerfile controller

build-worker:
	docker build -t $(WORKER_IMAGE) -f dockerfiles/controller.Dockerfile lettore