set -eu

sudo yum install -y htop git gcc cmake

git clone https://github.com/emscripten-core/emsdk
cd emsdk
./emsdk install latest
./emsdk activate latest
cd ..

git clone https://github.com/For-ACGN/keystone

# when use emscripten, it will cost a lot of memory.
#
# sudo fallocate -l 8G /swapfile
# sudo chmod 600 /swapfile
# sudo mkswap /swapfile
# sudo swapon /swapfile
#
# sudo swapoff /swapfile
# sudo rm /swapfile
