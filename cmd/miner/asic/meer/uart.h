#ifndef UART_H
#define UART_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <stdbool.h>
#include <termios.h>
#include <sys/ioctl.h>
#include <sys/types.h>
#include <errno.h>
#include <stdint.h>


#define UART_DBG_ENABLE     1

#define DEFAULT_UART "/dev/ttyS1"
#define DEFAULT_BAUDRATE  B1000000

int uart_open(char *devname, speed_t baud);
void uart_close(int fd);

bool uart_gets(int fd, uint8_t *buf, int r_size);
bool uart_write(int fd,  const void *buf, size_t wsize);

uint32_t uart_read_register(int fd, uint32_t chipid, uint32_t RegAddr);
unsigned int uart_write_register(int fd, uint32_t writemode,uint32_t burst_cn, uint32_t chipId, uint32_t Regaddr, uint32_t value);

void uart_set_host_baudrate(int fd, speed_t speed);

extern bool get_nonce(int fd, uint8_t *nonce, uint8_t *chip_id, uint8_t* job_id);

#endif
