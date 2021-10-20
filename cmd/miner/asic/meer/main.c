#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/time.h>

#include "uart.h"
#include "meer_drv.h"
#include "meer.h"
#include "main.h"
#define MEER_DRV_VERSION	"0.2asic"
#define NUM_OF_CHIPS	1
#define DEF_WORK_INTERVAL   30000 //msx


//测试主程序
int init_drv(int num_of_chips,char *path,char *gpio)
{
	int fd;
	printf("\n********************************Meer Driver %s - UART PATH:%s\n", MEER_DRV_VERSION,path);
	//初始化算力板
	if(meer_drv_init(&fd, num_of_chips,path,gpio)) {
		return -1;
	}

	meer_drv_set_freq(fd, 100);
	usleep(500000);
	uart_write_register(fd,0x90,0x00,0x00,0xff,0x00);   //门控
	usleep(100000);
	uart_write_register(fd,0x90,0x00,00,0x57,0x01);   //group 1
	usleep(100000);
	uart_write_register(fd,0x90,0x00,00,0x58,0x01);   //group 2
	usleep(100000);
	uart_write_register(fd,0x90,0x00,00,0x59,0x01);   //group 3
	usleep(100000);
	uart_write_register(fd,0x90,0x00,0x00,0xff,0x01);
	usleep(100000);
	uart_read_register(fd, 0x01, 0x00);
	uart_read_register(fd, 0x01, 0x57);
	uart_read_register(fd, 0x01, 0x58);
	uart_read_register(fd, 0x01, 0x59);

	meer_drv_set_freq(fd, 125);
	usleep(500000);
	meer_drv_set_freq(fd, 150);
	usleep(500000);
	meer_drv_set_freq(fd, 175);
	usleep(500000);
	meer_drv_set_freq(fd, 200);
	usleep(500000);
	meer_drv_set_freq(fd, 225);
	usleep(500000);
	meer_drv_set_freq(fd, 250);
	usleep(500000);
	return fd;
}

// 给固定芯片下发任务
void set_work(int fd,uint8_t* header,int pheader_len,uint8_t* target,int chipId,uint8_t* nonceStartA,uint8_t* nonceStartB,uint8_t* nonceStartC)
{
	struct work work_temp;
	memcpy(work_temp.target, target, 32); // 难度目标配置
	memcpy(work_temp.header, header, pheader_len); // meer区块头
	meer_drv_set_work(fd, &work_temp, chipId,nonceStartA,nonceStartB,nonceStartC); // 对算力板下任务
}
