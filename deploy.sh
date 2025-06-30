#!/bin/bash

LOCAL_PROJECT_DIR="." 
GO_SOURCE_FILE="sigma.go"
LINUX_EXECUTABLE_NAME="sigma_executable_linux"
FRONTEND_HTML_FILE="index.html"
RUN_SCRIPT_NAME="run.sh"
SSH_USER="t7ru"
SSH_HOST="ssh-t7ru.alwaysdata.net"
REMOTE_APP_DIR="/home/${SSH_USER}/sigma"

echo "Starting deployment script..."

cd "$LOCAL_PROJECT_DIR" || { echo "Failed to navigate to project directory: $LOCAL_PROJECT_DIR"; exit 1; }
echo "Current directory: $(pwd)"

if [ ! -f "$RUN_SCRIPT_NAME" ]; then
    echo "Error: $RUN_SCRIPT_NAME not found in $LOCAL_PROJECT_DIR."
    echo "Please create it. It should contain:"
    echo "#!/bin/sh"
    echo "cd \"\$(dirname \"\$0\")\""
    echo "chmod +x ./$LINUX_EXECUTABLE_NAME"
    echo "./$LINUX_EXECUTABLE_NAME"
    exit 1
fi

echo "Building Go application for Linux..."
GOOS=linux GOARCH=amd64 go build -v -o "$LINUX_EXECUTABLE_NAME" "$GO_SOURCE_FILE"
if [ $? -ne 0 ]; then
    echo "Go build failed!"
    exit 1
fi
echo "Build successful: $LINUX_EXECUTABLE_NAME created."

echo "Attempting to remove existing $LINUX_EXECUTABLE_NAME on the server..."
ssh "${SSH_USER}@${SSH_HOST}" "rm -f \"${REMOTE_APP_DIR}/${LINUX_EXECUTABLE_NAME}\""

echo "Uploading front.js to $SSH_USER@$SSH_HOST:$REMOTE_APP_DIR ..."
scp "front.js" "${SSH_USER}@${SSH_HOST}:${REMOTE_APP_DIR}/"
if [ $? -ne 0 ]; then
    echo "SCP upload failed for front.js!"
    exit 1
fi

echo "Uploading style.css to $SSH_USER@$SSH_HOST:$REMOTE_APP_DIR ..."
scp "style.css" "${SSH_USER}@${SSH_HOST}:${REMOTE_APP_DIR}/"
if [ $? -ne 0 ]; then
    echo "SCP upload failed for style.css!"
    exit 1
fi

echo "Uploading $LINUX_EXECUTABLE_NAME to $SSH_USER@$SSH_HOST:$REMOTE_APP_DIR ..."
scp "$LINUX_EXECUTABLE_NAME" "${SSH_USER}@${SSH_HOST}:${REMOTE_APP_DIR}/"
if [ $? -ne 0 ]; then
    echo "SCP upload failed for $LINUX_EXECUTABLE_NAME!"
    exit 1
fi

echo "Uploading $FRONTEND_HTML_FILE to $SSH_USER@$SSH_HOST:$REMOTE_APP_DIR ..."
scp "$FRONTEND_HTML_FILE" "${SSH_USER}@${SSH_HOST}:${REMOTE_APP_DIR}/"
if [ $? -ne 0 ]; then
    echo "SCP upload failed for $FRONTEND_HTML_FILE!"
    exit 1
fi

echo "Uploading $RUN_SCRIPT_NAME to $SSH_USER@$SSH_HOST:$REMOTE_APP_DIR ..."
scp "$RUN_SCRIPT_NAME" "${SSH_USER}@${SSH_HOST}:${REMOTE_APP_DIR}/"
if [ $? -ne 0 ]; then
    echo "SCP upload failed for $RUN_SCRIPT_NAME!"
    exit 1
fi
echo "Files uploaded successfully."

echo "Executing remote commands..."
ssh "${SSH_USER}@${SSH_HOST}" << EOF
    echo "Connected to remote server."
    cd "$REMOTE_APP_DIR" || { echo "Failed to navigate to remote directory: $REMOTE_APP_DIR"; exit 1; }
    echo "Current remote directory: \$(pwd)"

    echo "Setting execute permissions..."
    chmod +x "$LINUX_EXECUTABLE_NAME"
    chmod +x "$RUN_SCRIPT_NAME"
    echo "Permissions set."

    echo "Attempting to kill any old running process (if any)..."
    pkill -f "$REMOTE_APP_DIR/$LINUX_EXECUTABLE_NAME" || true 
    echo "Old process kill attempt finished (Alwaysdata should restart the new version)."

    echo "Remote commands finished."
EOF

if [ $? -ne 0 ]; then
    echo "SSH remote command execution failed!"
    exit 1
fi

echo "Deployment script finished successfully!"