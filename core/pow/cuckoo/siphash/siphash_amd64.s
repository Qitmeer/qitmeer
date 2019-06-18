//+build amd64 !noasm !appengine

#include "textflag.h"

/*
X0=v0
X1=v1
X2=v2
X3=v3
X4:tmp
X5:ff
X6:nonce
X7:rotate16
X8:X0
X9:X1
X10:X2
X11:X3
X12:uorv
*/

DATA rotate16<>+0x00(SB)/8, $0x0504030201000706
DATA rotate16<>+0x08(SB)/8, $0x0D0C0B0A09080F0E
GLOBL rotate16<>(SB), (NOPTR+RODATA), $16

#define ADD(a,b)\
	PADDQ	b, a

#define XOR(a, b)\
	PXOR	b, a

#define ROT(x,n) \
	MOVOA	x, X4 \
	PSLLQ	$n, X4 \ 
	PSRLQ	$(64-n), x \
	POR	X4, x 

#define ROT16(x) \
	PSHUFB	X7, x

#define ROT32(x) \
	PSHUFD	$0xB1, x, x          //X5 = X5[1,0,3,2]


#define SIPROUND \
    ADD(X0, X1) \
	ADD(X2, X3) \
	ROT(X1, 13) \
    ROT16(X3)\
	XOR(X1, X0)\
	XOR(X3, X2) \
    ROT32(X0) \
	ADD(X2, X1) \ 
	ADD(X0, X3) \
    ROT(X1, 17) \
	ROT(X3, 21) \
    XOR(X1, X2) \
	XOR(X3, X0) \
	ROT32(X2)


#define SIP1 \
	XOR(X3, X6) \
	SIPROUND  \
	SIPROUND  \
	XOR(X0, X6) \
	PXOR	X5, X2 \
	SIPROUND \
	SIPROUND \
	SIPROUND \
	SIPROUND \
	XOR(X0, X1) \
	XOR(X2, X3) \
	XOR(X0, X2)

#define NSIP \
	MOVOU	(CX),X6 \
	PSLLQ	$1, X6 \
	POR		X12,X6 \
	ADDQ	$16, CX \
	MOVOA	X8,X0 \ 
	MOVOA	X9,X1 \ 
	MOVOA	X10,X2 \
	MOVOA	X11,X3 \
	SIP1 \
	MOVOU X0, (BX) \
	ADDQ $16, BX

//func siphashPRF8192(v *[4]uint64, nonce *[8192]uint64, uorv uint64, result *[8192]uint64)
TEXT ·SiphashPRF8192(SB), NOSPLIT, $0
	MOVQ	$0xff, AX
	MOVQ	AX, X5
	MOVLHPS	X5, X5
	MOVOA	rotate16<>(SB),X7

	MOVQ	v+0(FP), CX
	MOVQ	(CX), X8
	ADDQ	$8, CX
	MOVQ	(CX), X9
	ADDQ	$8, CX
	MOVQ	(CX), X10
	ADDQ	$8, CX
	MOVQ	(CX), X11
    MOVQ	nonce+8(FP), CX
	MOVQ	uorv+16(FP), X12
	MOVQ	result+24(FP), BX
	MOVLHPS	X8, X8
	MOVLHPS	X9, X9 
	MOVLHPS	X10, X10 
	MOVLHPS	X11, X11
	MOVLHPS	X12, X12

	MOVQ $4096,R8
L1:
	NSIP //0
	DECQ R8
	JNZ L1
	RET

#define NSIP_SEQ \
	MOVOA	X8,X0 \ 
	MOVOA	X9,X1 \ 
	MOVOA	X10,X2 \
	MOVOA	X11,X3 \
	SIP1 \
	MOVOU X0, (BX) \
	ADDQ $16, BX \
	MOVQ	$4, AX \
	MOVQ	AX, X4 \
	MOVLHPS	X4, X4 \
	PADDQ	X4, X6 \


//func siphashPRF8192Seq(v *[4]uint64, nonce uint64, uorv uint64, result *[8192]uint64)
TEXT ·SiphashPRF8192Seq(SB), NOSPLIT, $0
	MOVQ	$0xff, AX
	MOVQ	AX, X5
	MOVLHPS	X5, X5
	MOVOA	rotate16<>(SB),X7

	MOVQ	v+0(FP), CX
	MOVQ	(CX), X8
	ADDQ	$8, CX
	MOVQ	(CX), X9
	ADDQ	$8, CX
	MOVQ	(CX), X10
	ADDQ	$8, CX
	MOVQ	(CX), X11
    MOVQ	nonce+8(FP), X6
	MOVQ	uorv+16(FP), X12
	MOVQ	result+24(FP), BX
	MOVLHPS	X8, X8
	MOVLHPS	X9, X9 
	MOVLHPS	X10, X10 
	MOVLHPS	X11, X11
	MOVLHPS	X12, X12

	MOVLHPS	X6, X6
	MOVQ	$1, AX
	MOVQ	AX, X4
	PSHUFD	$0x4e, X4, X4 //0x01001110   
	PADDQ	X4, X6
	PSLLQ	$1, X6
	POR		X12, X6

	MOVQ $4096,R8
L2:
	NSIP_SEQ //0
	DECQ R8
	JNZ L2
	RET

