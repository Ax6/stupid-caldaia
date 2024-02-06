APP_IMAGE=stupid-caldaia-app
TARGET_ADDRESS=192.168.1.123
TARGET_MACHINE=pi@$(TARGET_ADDRESS)
CONTROLLER_IMAGE=stupid-caldaia-controller
WORKER_IMAGE=stupid-caldaia-worker

# Typical deployment process
# make sync
# make bundle-client
# make bundle-server
# make restart-target
deploy: sync bundle-client bundle-server restart-target
deploy-client: sync bundle-client restart-target
deploy-server: sync bundle-server restart-target

init: # Don't forget about me!
	docker run --privileged --rm tonistiigi/binfmt --install all

run-app-target:
	cd app && PUBLIC_SERVER_HOST=$(TARGET_ADDRESS) PUBLIC_CLIENT_HOST=$(TARGET_ADDRESS) npm run dev && cd..

sync:
	git push && ssh $(TARGET_MACHINE) 'cd stupid-caldaia && git pull'

restart:
	docker compose stop && docker compose rm -f && docker compose up -d

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