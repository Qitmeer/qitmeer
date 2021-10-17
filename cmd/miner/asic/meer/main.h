#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "uart.h"
#include "meer_drv.h"
#include "meer.h"

extern int init_drv(int num_of_chips,char *path,char *gpio);
extern void set_work(int fd,uint8_t* header,int pheader_len,uint8_t* target,int chipId,uint8_t* nonceStartA,uint8_t* nonceStartB,uint8_t* nonceStartC);