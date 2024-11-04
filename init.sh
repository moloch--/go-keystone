set -eu

sudo yum install -y htop git gcc cmake

git clone https://github.com/emscripten-core/emsdk.git
cd emsdk
./emsdk install latest
./emsdk activate latest
source ./emsdk_env.sh
cd ..

git clone https://github.com/keystone-engine/keystone

# when use emscripten, it will cost a lot of memory.
#
# sudo fallocate -l 8G /swapfile
# sudo chmod 600 /swapfile
# sudo mkswap /swapfile
# sudo swapon /swapfile
#
# sudo swapoff /swapfile
# sudo rm /swapfile
