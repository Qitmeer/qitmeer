
In order to use [go-dfinity-crypto](https://github.com/dfinity/go-dfinity-crypto), need to build bn first.

### How to build bn From source

Your system needs have installed first llvm, g++, gmp and openssl{-dev}.

On Ubuntu:

```
sudo apt install llvm g++ libgmp-dev libssl-dev
```

On Mac

```
brew install llvm g++ libgmp-dev libssl-dev
```

To install from source, you can run the following commands:
```bash
git clone https://github.com/dfinity/bn
cd bn
make
make install
```
The library is then installed under `/usr/local/`.

environments setting

```bash
export LD_LIBRARY_PATH=/lib:/usr/lib:/usr/local/lib
```

