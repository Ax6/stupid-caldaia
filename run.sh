#!/bin/zsh
# This file runs together the following services for local development
# 1. redis-stack-server
# 2. svelte front-end in app/
# 3. control back-end at control/server.go
#
# However it doesn't run the worker locally as we assuming the hardware is not
# available you can emulate the worker by sending messages to the redis queue.
# 
# To execut this file make sure it's owned by the current user and has execute
# permissions. You can do this by running `chmod u+x run.sh` in the terminal.

# Make tmp folder where to pipe the logs
source ~/.zshrc
mkdir -p tmp

# Start redis
echo -n "💎 Starting redis... "
docker run -d --rm --name redis-caldaia -p 6379:6379 redis/redis-stack-server:latest > /dev/null
docker logs -f redis-caldaia > tmp/redis.log &
echo "OK"

# Start the control server
echo -n "🛂 Starting control service... "
{
cd controller && air server.go > ../tmp/control.log &
CONTROL_PID=$!
cd ..
} || {
	echo "🚨 Error: Failed to start control service"
	docker stop redis-caldaia > /dev/null
	exit 1
}
echo "OK"

# Start mock worker
echo -n "🤫 Starting mock work service... "
{
cd lettore/mock && CONFIG_PATH=../../config.json air lettore.go > ../tmp/lettore.log &
CONTROL_PID=$!
cd ../..
} || {
	echo "🚨 Error: Failed to start worker service"
	docker stop redis-caldaia > /dev/null
	exit 1
}
echo "OK"

# Start the svelte front-end
echo -n "🖥️  Starting svelte front-end... "
{
	cd app && PUBLIC_SERVER_HOST=localhost PUBLIC_CLIENT_HOST=localhost npm run dev > ../tmp/app.log &
} || {
	echo "🚨 Error: Failed to start svelte front-end"
	kill $CONTROL_PID
	docker stop redis-caldaia > /dev/null
	exit 1
}
APP_PID=$!
echo "OK"


# Wait for the user to press enter
echo "👌 All services started"
echo "Follow all logs: tail -f tmp/*.log"
echo "Redis: tail -f tmp/redis.log"
echo "Control: tail -f tmp/control.log"
echo "Worker: tail -f tmp/lettore.log"
echo "App: tail -f tmp/app.log"
echo "🟢 Press Enter to stop all services..."
read

# Kill the processes
kill $CONTROL_PID
kill $APP_PID
docker stop redis-caldaia > /dev/null
echo "All stopped 👋"

# Exit
exit 0