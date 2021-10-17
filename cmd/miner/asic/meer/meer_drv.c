#include <fcntl.h>


#include "uart.h"
#include "meer_drv.h"
#include "meer.h"

#define CHIP_CORE_TEST  0 //chip core self test

//算力芯片频率表,已计算好
struct miner_freq {
    uint32_t freq;
    uint32_t reg_value;
};


static const struct miner_freq miner_freqs[] = {
    {100,    0x00D82401},
    {125,    0x00B82581},
    {150,    0x00B82D01},
    {175,    0x00982A01},
    {200,    0x00942801},
    {225,    0x00942D01},
    {250,    0x00782D01},
    {275,    0x00742941},
    {300,    0x00742D01},
    {320,    0x00743001},
    {325,    0x00582701},
    {331, 0x005827C1},
    {337,  0x00582881},
    {343, 0x00582941},    
    {350,    0x00582A01},
    {356,    0x00582AC1},
    {362,    0x00582B81},
    {368,    0x00582C41},
    {375,    0x00582D01},
    {381,    0x00582DC1},
    {387,    0x00582E81},
    {393,    0x00582F41},    
    {400,    0x00542801},
    {425,    0x00542A81},
    {445,    0x00542C81},
    {447,    0x00542CC1},
    {450,    0x00542D01},
    {452,    0x00542D41},
    {455,    0x00542D81},
    {457,    0x00542DC1},
    {460,    0x00542E01},
    {462,    0x00542E41},
    {465,    0x00542E81},
    {467,    0x00542EC1},
    {470,    0x00542F01},
    {472,    0x00542F41},
    {475,    0x00542F81},
    {477,    0x00542FC1},
    {480,    0x00543001},
    {482,    0x00543041},
    {485,    0x00543081},
    {487,    0x005430C1},
    {490,    0x00543101},
    {492,    0x00543141},
    {495,    0x00543181},
    {496,    0x005027C1},
    {500,    0x00502801},
    {503,    0x00502841},
    {506,    0x00502881},
    {509,    0x005028C1},
    {512,    0x00502901},
    {515,    0x00502941},
    {518,    0x00502981},
    {521,    0x005029C1},    
    {525,    0x00502A01},
    {528,    0x00502A41},
    {531,    0x00502A81},
    {534,    0x00502AC1},
    {537,    0x00502B01},
    {540,    0x00502B41},
    {543,    0x00502B81},
    {546,    0x00502BC1},
    {550,    0x00502C01},
    {553,    0x00502C41},
    {556,    0x00502C81},
    {559,    0x00502CC1},
    {562,    0x00502D01},
    {565,    0x00502D41},
    {568,    0x00502D81},
    {571,    0x00502DC1},
    {575,    0x00502E01},
    {578,    0x00502E41},
    {581,    0x00502E81},
    {584,    0x00502EC1},
    {587,    0x00502F01},
    {590,    0x00502F41},
    {593,    0x00502F81},
    {596,    0x00502FC1},
    {600,    0x00503001},
    {604,    0x004C2441},
    {608,    0x004C2481},
    {612,    0x004C24C1},
    {616,    0x004C2501},
    {620,    0x004C2541},
    {625,    0x004C2581},
    {629,    0x004C25C1},
    {633,    0x004C2601},
    {637,    0x004C2641},
    {641,    0x004C2681},
    {645,    0x004C26C1},
    {650,    0x004C2701},
    {654,    0x004C2741},
    {658,    0x004C2781},
    {662,    0x004C27C1},
    {666,    0x004C2801},
    {670,    0x004C2841},
    {675,    0x004C2881},
    {679,    0x004C28C1},
    {683,    0x004C2901},
    {687,    0x004C2941},
    {691,    0x004C2981},
    {695,    0x004C29C1},
    {700,    0x004C2A01},
    {704,    0x004C2A41},
    {708,    0x004C2A81},
    {712,    0x004C2AC1},
    {716,    0x004C2B01},
    {720,    0x004C2B41},
    {725,    0x004C2B81},
    {729,    0x004C2BC1},
    {733,    0x004C2C01},
    {737,    0x004C2C41},
    {741,    0x004C2C81},
    {745,    0x004C2CC1},
    {750,    0x004C2D01},
    {775,    0x004C2E81},
    {800,    0x004C3001},
    {850,    0x00482201},
    {875,    0x00482301},
    {900,    0x00482401},
    {925,    0x00482501},
    {950,    0x00482601},
    {1000,   0x00482801},
    {1025,   0x00482901}

};

static uint32_t get_freq_reg_data(uint32_t freq)
{
    for(int i=0; i<sizeof(miner_freqs)/sizeof(struct miner_freq); i++) {
        if(miner_freqs[i].freq >= freq) {
            return miner_freqs[i].reg_value;
        }        
    }
    return 0x00742D01;  //300M
}

static void meer_drv_send_cmds(int fd, const unsigned char *cmds[])
{
	int i;
	for(i = 0; cmds[i] != NULL; i++)
	{
		uart_write(fd, cmds[i] + 1, cmds[i][0]);
		usleep(10000);
	}
}

/*auto_address*/
/*使能自动配芯片ID, 初始化必须, 给每个芯片配一个独立的ID, 从1依次递增*/
const unsigned char *cmd_auto_address[]={
	(unsigned char []) {0x08,0x90,0x00,0x00,0x80,0x01,0x00,0x00,0x00},
	NULL
};
/*配置直通模式， 初始化必须*/
const unsigned char *cmd_feedthr_clear_slot[]={//配置feed through模式
	(unsigned char []) {0x08,0x90,0x00,0x00,0x81,0x03,0x00,0x00,0x00},
	NULL
};

#define RST0_VAL "/sys/class/gpio/gpio128/value"
enum {
    CMD_WRITE = -1, CMD_READ = -2
};
#define GPIO_HIGH	"1"
#define GPIO_LOW	"0"    
static inline int common_cmd(int cmd, char *node, char *value){
	int fd = -1;
	char buf[10] = {0};

	switch(cmd){
		case CMD_WRITE:
			fd = open(node, O_WRONLY);
            if(fd < 0) {
                perror(strerror(errno));
                return (-1);
            }
			write(fd, value, 10);
			close(fd);			
			return 0;
		break;
		case CMD_READ:
			fd = open(node, O_RDONLY);
            if(fd < 0) {
                perror(strerror(errno));
                return (-1);
            }
			read(fd, buf, 10);
			close(fd);			
			return atoi(buf);
		break;
		default:
			return -1;
		break;
	}
    return -1;
}

void meer_drv_reset_pin(uint8_t value, bool reset,char *gpio)
{
    char *rst_val[4] = {gpio};
	
    if(reset) {
        //TODO:
        common_cmd(CMD_WRITE, rst_val[0], GPIO_LOW); usleep(300000);
    	common_cmd(CMD_WRITE, rst_val[0], GPIO_HIGH); usleep(300000);
    	common_cmd(CMD_WRITE, rst_val[0], GPIO_LOW); usleep(300000);
    } else {
        //TODO:
        if(value) {
            common_cmd(CMD_WRITE, rst_val[0], GPIO_HIGH);
        } else {
            common_cmd(CMD_WRITE, rst_val[0], GPIO_LOW);
        }
    }
}

bool meer_drv_init(int *fd, int num_chips,char *path,char *gpio)
{
    int fdtemp;
    meer_drv_reset_pin(0, true,gpio);
    
    fdtemp = uart_open(path, DEFAULT_BAUDRATE);
    *fd = fdtemp;
    if(*fd < 0) {
        return false;
    }

    if (!uart_write_register(fdtemp,0x90,0x00,0x00,0x81,0x00)){
        return false;
    }	//进入配芯片ID模式, 次序不能动
    usleep(100000);
    meer_drv_send_cmds(fdtemp, cmd_auto_address);			//配芯片ID, 次序不能动
    usleep(500000);
    uart_write_register(fdtemp,0x90,0x00,0x00,0x81,0x01);	//退出配芯片ID模式
    usleep(100000);
    uart_write_register(fdtemp,0x90,0x00,0x00,0x82,DEF_SLOT_DEFAULT*num_chips);	//配芯片发送总时隙
    usleep(10000);
    for(int i=1; i<(num_chips+1); i++)
		uart_write_register(fdtemp,0x44,0x00,i,0x83,DEF_SLOT_DEFAULT*(i-1));//给每个芯片配独立的发送时隙
    usleep(100000);
    meer_drv_send_cmds(fdtemp, cmd_feedthr_clear_slot); //进入挖矿模式
    usleep(100000);
       
    if(uart_read_register(fdtemp, 0x01, 0x00) != 0xaa) { //读芯片ID
        return false;
    }
    
    return true;
}
void meer_drv_deinit(int fd,char *gpio)
{    
    close(fd);
    meer_drv_reset_pin(0, false,gpio);
}

void meer_drv_set_freq(int fd, uint32_t freq)
{
    printf("\n********************set freq %d\n",freq);
    uart_write_register(fd, 0x90, 0, 0, 0xf3, 0x2f);
    uart_write_register(fd, 0x90, 0, 0, 0xf0, 0x00);
    uart_write_register(fd, 0x90, 0, 0, 0xf1, get_freq_reg_data(freq));            
    uart_write_register(fd, 0x90, 0, 0, 0xf3, 0x2e);
}

static const unsigned char *cmd_soft_reset[]={
	(unsigned char []) {0x08,0x90,0x00,0x00,0x81,0x03,0x00,0x00,0x00},
	(unsigned char []) {0x08,0x90,0x00,0x00,0xff,0x00,0x00,0x00,0x00},
	(unsigned char []) {0x08,0x90,0x00,0x00,0xff,0x07,0x00,0x00,0x00},
	NULL
};

//芯片软复位
void meer_drv_softreset(int fd)
{
    meer_drv_send_cmds(fd, cmd_soft_reset);
}

//给芯片下任务
//参数work: 要计算的任务
//参数num_chips: 给多少颗芯片下任务
bool meer_drv_set_work_old(int fd, struct work *work, int num_chips)
{
    uint8_t chip_id = 1;
    uint64_t index = 0;
    uint64_t unit = 0xffffffffffffff; // 7个字节

    for(;chip_id<=num_chips;chip_id++) {
        unsigned char midstate[256]={0};
        int midstate_len = 0;
        unsigned char bin[256]={0};
        unsigned int force_start = 0;
        unsigned char* pstart = (unsigned char*)(&force_start);
        int bpos = 0;
        uint8_t group_id = 0;
        uint8_t jobid = 0;
       
        memcpy(bin, "\x44\x01\x00\x00", 4);
        bpos = 4;
        
        memcpy(bin + bpos, &(work->target[24]), 8);
        bpos += 8;

        midstate_len = meer_calc_midstate(midstate, work->header); //midstate, 区块头头算出的中间结果
        memcpy(bin + bpos, midstate, midstate_len);        
        bpos += midstate_len;

        memcpy(bin + bpos, &(work->header[72]), 36+9);
        bpos += 48;

        bin[1] = (bpos-4)/4-1; //data size
        bin[2] = chip_id; //chip id
        for(group_id = 0; group_id < DEF_CHIP_MAX_GROUPS; group_id++) {
            uint64_t step = unit * index;
            index++;
            uint8_t *p = (uint8_t *)&step;
//            if(1 == group_id) {
//                memcpy(&(bin[bpos-8]), "\x55\x55\x55\x55\x55\x55\x55\x55", 8); //init nonce range 1
//            } else if(2 == group_id) {
//                memcpy(&(bin[bpos-8]), "\xaa\xaa\xaa\xaa\xaa\xaa\xaa\xaa", 8); //init nonce range 2
//            } else {
//                memcpy(&(bin[bpos-8]), "\x00\x00\x00\x00\x00\x00\x00\x00", 8); //init nonce range 0
//            }
            memcpy(&(bin[bpos-8]), p, 8); //init nonce range 0
            uart_write(fd, bin, bpos);
            
            uart_write_register(fd,0x44,0x00,chip_id,0x40, 0xf1818001);            
            uart_write_register(fd,0x44,0x00,chip_id,0x42, 1<<group_id);
            
#if CHIP_CORE_TEST
                static bool g_core_test = false;
                if(!g_core_test) {
                    g_core_test = true;
                    uart_write_register(fd, 0x90, 0x00, 0x00, 0x41, 0x00010fc0);
                }
#else 

            force_start = 0;        
            
            if(group_id < 2) {
                pstart[0] = 1<<(group_id+6);
            } else {
                pstart[1] = 1<<(group_id-2);
            }            
            pstart[1] += (jobid<<4);    
        	uart_write_register(fd,0x44,0x00,chip_id,0x41,force_start); //group1
            printf("\n%s, chip %d, group %d, job %d nonce start %llu \n", __func__, chip_id, group_id, jobid,step);
            jobid++;            
        }
    }
#endif    
}

//给芯片下任务
//参数work: 要计算的任务
//参数num_chips: 给多少颗芯片下任务
bool meer_drv_set_work(int fd, struct work *work, int chip_id,uint8_t* nonceStartA,uint8_t* nonceStartB,uint8_t* nonceStartC)
{
	    unsigned char midstate[256]={0};
	    int midstate_len = 0;
	    unsigned char bin[256]={0};
	    unsigned int force_start = 0;
	    unsigned char* pstart = (unsigned char*)(&force_start);
	    int bpos = 0;
	    uint8_t group_id = 0;
	    uint8_t jobid = 0;

	    memcpy(bin, "\x44\x01\x00\x00", 4);
	    bpos = 4;

	    memcpy(bin + bpos, &(work->target[24]), 8);
	    bpos += 8;

	    midstate_len = meer_calc_midstate(midstate, work->header); //midstate, 区块头头算出的中间结果
	    memcpy(bin + bpos, midstate, midstate_len);
	    bpos += midstate_len;

	    memcpy(bin + bpos, &(work->header[72]), 36+9);
	    bpos += 48;

	    bin[1] = (bpos-4)/4-1; //data size
	    bin[2] = chip_id; //chip id
	    for(group_id = 0; group_id < DEF_CHIP_MAX_GROUPS; group_id++) {
            if(1 == group_id) {
                memcpy(&(bin[bpos-8]), nonceStartA, 8); //init nonce range 1
            } else if(2 == group_id) {
                memcpy(&(bin[bpos-8]), nonceStartB, 8); //init nonce range 2
            } else {
                memcpy(&(bin[bpos-8]), nonceStartC, 8); //init nonce range 0
            }
            uart_write(fd, bin, bpos);

            uart_write_register(fd,0x44,0x00,chip_id,0x40, 0xf1818001);
            uart_write_register(fd,0x44,0x00,chip_id,0x42, 1<<group_id);

        #if CHIP_CORE_TEST
                static bool g_core_test = false;
                if(!g_core_test) {
                    g_core_test = true;
                    uart_write_register(fd, 0x90, 0x00, 0x00, 0x41, 0x00010fc0);
                }
        #else

            force_start = 0;

            if(group_id < 2) {
                pstart[0] = 1<<(group_id+6);
            } else {
                pstart[1] = 1<<(group_id-2);
            }
            pstart[1] += (jobid<<4);
            uart_write_register(fd,0x44,0x00,chip_id,0x41,force_start); //group1
            //printf("\n%s, chip %d, group %d, job %d \n", __func__, chip_id, group_id, jobid);
            jobid++;
        #endif
    }
}

