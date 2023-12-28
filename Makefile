APP_IMAGE=stupid-caldaia-app
TARGET_MACHINE=pi@192.168.1.112
CONTROLLER_IMAGE=stupid-caldaia-controller
WORKER_IMAGE=stupid-caldaia-worker

restart:
	docker-compose restart

restart-target:
	ssh $(TARGET_MACHINE) 'cd stupid-caldaia && make restart'

bundle-client: build-app transfer-app

bundle-server: build-executables transfer-executables
	ssh $(TARGET_MACHINE) 'cd stupid-caldaia && make docker-build-controller'
	ssh $(TARGET_MACHINE) 'cd stupid-caldaia && make docker-build-worker'

build-app:
	docker buildx build --platform=linux/arm64 -t $(APP_IMAGE) -f dockerfiles/app.Dockerfile app

transfer-app:
	docker save $(APP_IMAGE) | bzip2 | pv | ssh $(TARGET_MACHINE) 'bunzip2 | docker load'

docker-build-controller:
	docker buildx build --build-context executables=/home/pi/bin/stupid-caldaia -t $(CONTROLLER_IMAGE) -f dockerfiles/controller.Dockerfile .

docker-build-worker:
	docker buildx build --build-context executables=/home/pi/bin/stupid-caldaia -t $(WORKER_IMAGE) -f dockerfiles/worker.Dockerfile .

build-executables:
	cd controller && GOOS=linux GOARCH=arm64 go build -o controller && cd ..
	cd lettore && GOOS=linux GOARCH=arm64 go build -o lettore && cd ..

transfer-executables:
	ssh $(TARGET_MACHINE) "mkdir -p /home/pi/bin/stupid-caldaia"
	scp controller/controller $(TARGET_MACHINE):/home/pi/bin/stupid-caldaia/controller && rm controller/controller
	scp lettore/lettore $(TARGET_MACHINE):/home/pi/bin/stupid-caldaia/lettore && rm lettore/lettore