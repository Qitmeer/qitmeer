#include <fcntl.h>

#include "uart.h"

bool uart_gets(int fd, uint8_t *buf, int r_size)
{
    ssize_t nread = 0;
    int repeat = 0;
    int iocount = 0;
    int len = r_size;
    int total = 0;

    ioctl(fd, FIONREAD, &iocount);
    if (iocount < r_size)
    {
        return false;
    }
    memset(buf, 0, r_size);

    while(len > 0)
    {
        nread = read(fd, buf+total, len);
        if(nread < 0)
        {
            printf("%s Read error: %s\n", __func__, strerror(errno));
            return false;
        }

        len -= nread;
        total+=nread;

        if ((repeat++ > 1) && len)
        {
            printf("%s read failed\n", __func__);
            return false;
        }
        else
        {
            if(len > 0)
            {
                printf("%s, repeat\n", __func__);		
            }
        }
    }

    return true;
}

bool get_nonce(int fd, uint8_t *nonce, uint8_t *chip_id, uint8_t* job_id)
{
	#define MAGIC_HEADER 0xcc
    #define PACKET_LEN 11
    uint8_t buffer[24]={0};
    
    int size = 0, nonce_packet_size = PACKET_LEN;
    size = nonce_packet_size;

    uint8_t *pbuf = buffer;
    bool rd_ret = false;
	memset(buffer, 0, sizeof(buffer));
    rd_ret = uart_gets(fd, pbuf, size);

    if(rd_ret > 0) {
        /**
        printf("return:");
        for(int i=0;i<nonce_packet_size;i++) {
            printf("%02x", buffer[i]);
        }
        printf("\n"); **/
        if (buffer[0] == MAGIC_HEADER) {            
            memcpy((uint8_t *)(nonce), buffer+3, 8); //nonce: 8 bytes
            *chip_id = buffer[1]; //chip id: start from 1
            *job_id = buffer[2]&0x0f; //job id: 0~15 total 4bits           
            //printf("\n%s chip_id %d, job_id %d, %02x%02x%02x%02x%02x%02x%02x%02x\n", __func__, buffer[1], buffer[2]&0x0f, nonce[0], nonce[1], nonce[2], nonce[3], nonce[4], nonce[5], nonce[6], nonce[7]);
			return true;
        }
        
    }
    return false;
}


bool uart_write(int fd,  const void *buf, size_t wsize)
{
	int repeat = 0;
	int size = 0;
	int ret = 0;
	int nwrite = 0;
	int len = wsize;
	int total = 0;

	while(len > 0)
	{
		nwrite = write(fd, buf+total, len);		
		if (nwrite < 0)
		{
			printf("%s Write error: %s\n", __func__, strerror(errno));
			return false;
		}

		len -= nwrite;
		total += nwrite;
        fsync(fd);
		if ((repeat++ > 1) && len)
		{
            printf("%s uart write failed\n", __func__);
			return false;
		}
		else
		{
		    if(len > 0)
			{
				printf("%s, repeat\n", __func__);
			}
		}
		
	}
	fsync(fd);
	return true;
}

unsigned int uart_write_register(int fd, uint32_t writemode,uint32_t burst_cn, uint32_t chipId, uint32_t Regaddr, uint32_t value)
{
	bool ret =true;
	uint8_t read_reg_cmd[16]={0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00};
	read_reg_cmd[0] = writemode;
	read_reg_cmd[1] = burst_cn;
	read_reg_cmd[2] = chipId;
	read_reg_cmd[3] = Regaddr;
	read_reg_cmd[4] = value&0xff;
	read_reg_cmd[5] = (value>>8)&0xff;
	read_reg_cmd[6] = (value>>16)&0xff;
	read_reg_cmd[7] = (value>>24)&0xff;
	
	ret = uart_write(fd, read_reg_cmd, 8);

	return  ret;
}

uint32_t uart_read_register(int fd, uint32_t chipid, uint32_t RegAddr)
{
	bool ret = false;
	uint8_t read_reg_data[7]={0};
	uint8_t read_reg_cmd[10]={0x55, 0x00, 0x00, 0x00,0x00,0x00,0x00,0x00};
    uint32_t data;
    uint32_t* pdata = (uint32_t*)(&(read_reg_data[3]));
	read_reg_cmd[2] = (uint8_t)chipid;
	read_reg_cmd[3] = (uint8_t)RegAddr;
	
	uart_write(fd, read_reg_cmd, 8); 
	usleep(100000);
	if(uart_gets(fd, read_reg_data, 11))
	{
#if UART_DBG_ENABLE	
		printf("%s read addr 0x%02x return:", __func__, RegAddr);
		for (int i=0; i<7; i++)
		{
			printf("0x%02x ", read_reg_data[i]);
		}
		printf("\n");
#endif        
	}

    data = *pdata;

	return data;
}

int uart_open(char *devname, speed_t baud)
{

    struct termios	my_termios;
    int uart_fd = -1;

    do
    {
        uart_fd = open(devname, O_RDWR | O_CLOEXEC | O_NOCTTY);
        if (uart_fd == -1) {
            if (errno == EACCES) {                
                printf("%s Do not have user privileges to open %s\n", __func__, devname);
            }
            else {                
                printf("%s failed open device %s\n", __func__, devname);
            }
            return  -1;
        }

        if(tcgetattr(uart_fd, &my_termios)) break;
        if(cfsetispeed(&my_termios, baud)) break;
        if(cfsetospeed(&my_termios, baud)) break;        

        my_termios.c_cflag &= ~(CSIZE | PARENB | CSTOPB);
        my_termios.c_cflag |= CS8;
        my_termios.c_cflag |= CREAD;
        my_termios.c_cflag |= CLOCAL;

        my_termios.c_iflag &= ~(IGNBRK | BRKINT | PARMRK |
                        ISTRIP | INLCR | IGNCR | ICRNL | IXON);
        my_termios.c_oflag &= ~OPOST;
        my_termios.c_lflag &= ~(ECHO | ECHOE | ECHONL | ICANON | ISIG | IEXTEN);

        // Code must specify a valid timeout value (0 means don't timeout)
        my_termios.c_cc[VTIME] = (cc_t)10;
        my_termios.c_cc[VMIN] = 0;

        if(tcsetattr(uart_fd, TCSANOW, &my_termios)) break;
        if(tcflush(uart_fd, TCIOFLUSH)) break;
    }while(0);

    return uart_fd;
}

void uart_close(int fd)
{    
    close(fd);
}

void uart_set_host_baudrate(int fd, speed_t speed)
{
	struct termios my_termios;

	tcgetattr(fd, &my_termios);
	cfsetispeed(&my_termios, speed);
	cfsetospeed(&my_termios, speed);

	tcsetattr(fd, TCSANOW, &my_termios);
	tcflush(fd, TCIOFLUSH);
}



