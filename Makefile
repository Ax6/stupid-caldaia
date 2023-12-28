TARGET_ADDRESS=raspberrypi.local
TARGET_SHH=pi@raspberrypi
APP_IMAGE_NAME=stupid-caldaia
DOCKER_REGISTRY=$(TARGET_ADDRESS):5000

APP_IMAGE = $(DOCKER_REGISTRY)/$(APP_IMAGE_NAME)

bundle:
	docker build -t $(APP_IMAGE) --platform=linux/arm/v7 -f dockerfiles/app.Dockerfile app


# Since my company blocks everything we have to push manually the image via ssh to the target machine
# Make sure pv is installed first: sudo apt-get install pv
push:
	docker push $(APP_IMAGE)


# We need to host a registry on the target machine
target-dependencies:
	ssh $(TARGET_SHH) "sudo docker run -d -p 5000:5000 --restart=always --name registry registry"