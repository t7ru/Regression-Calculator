cd "$(dirname "$0")"

echo "Attempting to start pre-compiled application: sigma_executable_linux on port $PORT"
chmod +x ./sigma_executable_linux

./sigma_executable_linux