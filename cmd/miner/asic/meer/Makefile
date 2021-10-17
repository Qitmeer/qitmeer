SRCS := $(wildcard *.c)
CC := arm-linux-gnueabihf-gcc
OUTPUT_FILE := test_meer
all:
	$(CC) -O2 $(SRCS) -o $(OUTPUT_FILE)

clean:
	rm $(OUTPUT_FILE)
