#ifndef MEER_DRV_H
#define MEER_DRV_H

#define DEF_SLOT_1M     0xd00
#define DEF_SLOT_DEFAULT DEF_SLOT_1M

#define DEF_CHIP_MAX_GROUPS     3
#define DEF_CHIP_MAX_CORES      8

//给芯片下的任务结构
struct work {
	unsigned char target[32];	//计算目标值
    unsigned char header[117];	//区块头
};

bool meer_drv_init(int *fd, int num_chips,char *path,char *gpio); //算力板初始化
extern void meer_drv_deinit(int fd,char *gpio);

extern void meer_drv_set_freq(int fd, uint32_t freq);	//配置算力芯片频率
bool meer_drv_set_work(int fd, struct work *work, int num_chips,uint8_t* nonceStartA,uint8_t* nonceStartB,uint8_t* nonceStartC); //对算力芯片下计算任务

void meer_drv_reset_pin(uint8_t value, bool reset,char *gpio); //算力板复位

/*指令格式
注: 串口默认使用1Mbps波特率
注:B指字节
注:传输为小端顺序, 低字节在前，高字节在后
注: fpga芯片不用配频率, asic芯片需要配频率
下发指令:
注:burst count：指要下发的data字节数, =字节数/4-1， 如发送4字节data即为0， 8字节data即为1

广播写寄存器指令: 对所有芯片下发, chip id 为0
0x90(1B) + burst count(1B) + chip id(1B) + reg address(1B) + data
单播写寄存器指令: 对单个芯片下发，chip id 为对应芯片ID, 范围1~n
0x44(1B) + burst count(1B) + chip id(1B) + reg address(1B) + data

读寄存器指令:
0x55(1B) + burst count(1B) + chip id(1B) + reg address(1B)
读寄存器返回:
0xaa(1B) + chip id(1B) + reg address(1B) + data(4B)

下任务:
格式同写寄存器指令
job_id 用于区分下发/上报的是哪一个job， 范围0~15
group_id 用于指定发给哪一组计算core, 每个芯片共3个DEF_CHIP_MAX_GROUPS个group
job_id和group_id 见meer_drv_set_work函数

nonce返回:
0xcc(1B) + chip_id(1B) + job_id(1B,低4位有效，高4位忽略) + nonce(8B)
*/
#endif
