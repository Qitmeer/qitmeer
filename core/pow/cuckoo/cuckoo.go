package cuckoo

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/cuckoocycle/siphash"
	"github.com/pkg/errors"
	"runtime"
	"sort"
	"sync"
	//"github.com/btcsuite/btcd/wire"
	"qitmeer/core/types"
)

const (
	edgebits = 24 //the 2-log of the graph size,which is the size in bits of the node identifiers. the size of the edge indices in the Cuckoo Cycle graph.
	//edgebits  = 29
	ProofSize = 20 //ProofSize is the number of nonces and cycles. the (even) length of the cycle to be found. a minimum of 12 is recommended.
	//ProofSize = 42
	nedge    = 1 << edgebits    //number of edges：2的“edgebits”次方
	edgemask = nedge - 1        // used to mask siphash output
	nnode    = 2 * nedge        // 节点数是边数的2倍
	Easiness = nnode * 50 / 100 // 一个easiness值（a consensus value），它代表M/N，用来描述发现一个正确解的概率, Cuckoo Cycle固定M=N/2
	maxpath  = 8192
	//maxpath   = 8 << ((edgebits + 2)/3) //2048

	xbits       = 5
	comp0bits   = 32 - edgebits
	comp1bits   = 6
	xmask       = 0x1f
	zmask       = 0x3fff
	comp0mask   = 0xff
	ycomp0mask  = 0x1fff  // 5+8
	xymask      = 0x3ff   // 5+5
	xycomp1mask = 0x1ffff //5+5+6+1

	nx     = 1 << xbits
	zbits  = edgebits - 2*xbits
	nz     = 1 << zbits // 2^14,16384
	bigeps = nz + nz*3/64
)

type bucket []uint64

//Cuckoo is struct for cuckoo miner.
type Cuckoo struct {
	cuckoo []uint32
	sip    *siphash.SipHash
	matrix [nx][nx]bucket //stores all edges in a matrix of NX * NX buckets, where NX = 2^XBITS is the number of possible values of the 'X' bits.
	ncpu   int
	m2     [][nx]bucket

	us      []uint32
	vs      []uint32
	matrixs [][nx][nx]bucket
	m2tmp   [][nx]bucket
}

type edges struct {
	edge   []uint64
	uxymap []bool
}

//NewCuckoo returns Cuckoo struct to do PoW.
func NewCuckoo() *Cuckoo {
	ncpu := runtime.NumCPU()
	//fmt.Printf("maxpath = %v\n",maxpath)
	c := &Cuckoo{
		cuckoo:  make([]uint32, 1<<17+1),    // 创建length为2^17 + 1 的uint32数组,131073
		ncpu:    ncpu,                       // CPU核数，华为MateBook D 是 8核
		us:      make([]uint32, 0, maxpath), // 创建length为maxpath的uint32数组，并返回length为0，capacity为maxpath的slice
		vs:      make([]uint32, 0, maxpath),
		matrixs: make([][nx][nx]bucket, ncpu), //这里的nx=2^5,即32
		m2tmp:   make([][nx]bucket, ncpu),
	}
	if c.ncpu > 32 {
		c.ncpu = 32
	}
	c.m2 = make([][nx]bucket, c.ncpu) // 限定32列，但行数不定的二维数组
	// 初始化8行32列的二维数组，二维数组的每个元素又是一个uint64类型的数组
	for i := 0; i < c.ncpu; i++ {
		for x := 0; x < nx; x++ {
			c.m2[i][x] = make(bucket, 0, bigeps)
		}
	}
	// 初始化32行32列的二维数组，二维数组的每个元素又是一个uint64类型的数组
	for x := 0; x < nx; x++ {
		for y := 0; y < nx; y++ {
			c.matrix[x][y] = make([]uint64, 0, bigeps)
		}
	}
	// 初始化8行32列32高的三维数组，三维数组的每个元素又是一个uint64类型的数组
	for j := 0; j < c.ncpu; j++ {
		for x := 0; x < nx; x++ {
			for y := 0; y < nx; y++ {
				c.matrixs[j][x][y] = make(bucket, 0, bigeps/c.ncpu)
			}
		}
	}
	// 初始化8行32列的二维数组，二维数组的每个元素又是一个uint64类型的数组
	for j := 0; j < c.ncpu; j++ {
		for i := range c.m2tmp[j] {
			c.m2tmp[j][i] = make([]uint64, 0, bigeps)
		}
	}
	//TODO，测试用打印 c
	//fmt.Println("c.cuckoo = ",c.cuckoo)
	//fmt.Println("c.ncpu = ",c.ncpu)
	//fmt.Println("c.us = ",c.us)
	//fmt.Println("c.vs = ",c.vs)
	//fmt.Println("c.matrixs = ",c.matrixs)
	//fmt.Println("c.matrixs[0] type:", reflect.TypeOf(c.matrixs[0]))
	//fmt.Println("c.matrixs[0][0] type:", reflect.TypeOf(c.matrixs[0][0]))
	//fmt.Println("c.matrixs[0][0][0] type:", reflect.TypeOf(c.matrixs[0][0][0]))
	//fmt.Println("c.m2tmp = ",c.m2tmp)
	//fmt.Println("c.sip = ",c.sip)
	//fmt.Println("c.m2 = ",c.m2)

	return c
}

//PoW does PoW with hash, which is the key for siphash.
func (c *Cuckoo) PoW(siphashKey []byte) ([]uint32, bool) {
	// c.m2 = [1][8][32]uint64, 创建这样一个多维切片，但没有初始化，[[[[] [] [].32.].8.].1.]
	for i := 0; i < c.ncpu; i++ {
		for x := 0; x < nx; x++ {
			c.m2[i][x] = c.m2[i][x][:0]
		}
	}
	//fmt.Printf("PoW() c.m2 = %v\n",c.m2)

	// c.matrix = [1][32][32]uint64, 创建这样一个多维切片，但没有初始化，[[[[] [] [].32.].32.].1.]
	for x := 0; x < nx; x++ {
		for y := 0; y < nx; y++ {
			c.matrix[x][y] = c.matrix[x][y][:0]
		}
	}
	//fmt.Printf("PoW() c.matrix = %v\n",c.matrix)

	// 一维数组初始化为0，一共有131073(2^17+1)个元素被初始化为0
	for i := range c.cuckoo {
		c.cuckoo[i] = 0
	}

	//
	//fmt.Printf("siphashKey = %v\n",siphashKey)
	// siphashKey = [196 107 38 219 80 75 209 213 243 49 219 252 101 35 20 105]
	// s.k0 = 15407178610857372612
	// s.k1 = 7571715794457539059
	// s.V = [12015082867820662449 971459224712208030 13377991302290020773 2121554997101417600]
	c.sip = siphash.Newsip(siphashKey)

	//fmt.Printf("before buildU() c.matrix = %v\n",c.matrix)
	c.buildU()
	//fmt.Printf("after buildU() c.matrix = %v\n",c.matrix)
	c.buildV()
	//fmt.Printf("after buildV() c.matrix = %v\n",c.matrix)

	c.trimmimng()

	for _, ux := range c.matrix {
		for _, m := range ux {
			for _, uv := range m {
				u := uint32(uv>>32) << 1
				v := (uint32(uv) << 1) | 1
				us, err1 := c.path(u, c.us)
				vs, err2 := c.path(v, c.vs)
				if err1 != nil || err2 != nil {
					continue
				}
				if us[len(us)-1] == vs[len(vs)-1] {
					if ans, ok := c.solution(us, vs); ok {
						return ans, true
					}
					continue
				}
				if len(us) < len(vs) {
					for nu := len(us) - 2; nu >= 0; nu-- {
						c.cuckoo[us[nu+1]&xycomp1mask] = us[nu]
					}
					c.cuckoo[u&xycomp1mask] = v
				} else {
					for nv := len(vs) - 2; nv >= 0; nv-- {
						c.cuckoo[vs[nv+1]&xycomp1mask] = vs[nv]
					}
					c.cuckoo[v&xycomp1mask] = u
				}
			}
		}
	}
	return nil, false
}

//Verify cuckoo nonces.
func Verify(sipkey []byte, nonces [ProofSize]uint32) error {
	sip := siphash.Newsip(sipkey)
	var uvs [2 * ProofSize]uint32
	var xor0, xor1 uint32

	if len(nonces) != ProofSize {
		return errors.New("length of nonce is not correct")
	}

	if nonces[ProofSize-1] > Easiness {
		return errors.New("nonce is too big")
	}

	for n := 0; n < ProofSize; n++ {
		if n > 0 && nonces[n] <= nonces[n-1] {
			fmt.Printf("n=%d\n", n)
			return errors.New("nonces are not in order")
		}
		u00 := siphash.SiphashPRF(&sip.V, uint64(nonces[n]<<1))
		v00 := siphash.SiphashPRF(&sip.V, (uint64(nonces[n])<<1)|1)
		u0 := uint32(u00&edgemask) << 1
		xor0 ^= u0
		uvs[2*n] = u0
		v0 := (uint32(v00&edgemask) << 1) | 1
		xor1 ^= v0
		uvs[2*n+1] = v0
	}
	if xor0 != 0 {
		return errors.New("U endpoinsts don't match")
	}
	if xor1 != 0 {
		return errors.New("V endpoinsts don't match")
	}

	n := 0
	for i := 0; ; {
		another := i
		for k := (i + 2) % (2 * ProofSize); k != i; k = (k + 2) % (2 * ProofSize) {
			if uvs[k] == uvs[i] {
				if another != i {
					return errors.New("there are branches in nonce")
				}
				another = k
			}
		}
		if another == i {
			return errors.New("dead end in nonce")
		}
		i = another ^ 1
		n++
		if i == 0 {
			break
		}
	}
	if n != ProofSize {
		return errors.New("cycle is too short")
	}
	return nil
}

func (c *Cuckoo) trimmimng() {
	var i int

	_, maxv := c.trim(false)
	_, maxu := c.trim(true)
	for i = 3; maxu > 1<<(comp0bits+1) || maxv > 1<<(comp0bits+1); i += 2 {
		_, maxv = c.trim(false)
		_, maxu = c.trim(true)
	}
	c.trimrename0(false)
	c.trimrename0(true)
	for i += 2; i < 65; i += 2 {
		c.trim2(false)
		c.trim2(true)
	}
	c.trimrename1(false)
	c.trimrename1(true)
}

func (c *Cuckoo) path(u uint32, us []uint32) ([]uint32, error) {
	us = us[:0]
	nu := 0
	for ; u != 0; nu++ {
		if nu >= maxpath {
			return nil, errors.New("more than maxpath")
		}
		us = append(us, u)
		u = c.cuckoo[u&xycomp1mask]
	}
	return us, nil
}

func (c *Cuckoo) solution(us []uint32, vs []uint32) ([]uint32, bool) {
	nu := int32(len(us) - 1)
	nv := int32(len(vs) - 1)
	min := nu
	if min > nv {
		min = nv
	}
	nv -= min
	nu -= min
	for us[nu] != vs[nv] {
		nu++
		nv++
	}
	l := nu + nv + 1
	if l != ProofSize {
		return nil, false
	}

	es := newedges()
	es.add(us[0], vs[0])
	for nu--; nu >= 0; nu-- {
		es.add(us[(nu+1)&^1], us[nu|1])
	}
	for nv--; nv >= 0; nv-- {
		es.add(vs[nv|1], vs[(nv+1)&^1])
	}
	sort.Slice(es.edge, func(i, j int) bool {
		return es.edge[i] < es.edge[j]
	})
	answer := make([]uint32, 0, ProofSize)
	steps := Easiness / c.ncpu
	remain := Easiness - steps*c.ncpu
	var wg sync.WaitGroup
	var mutex sync.Mutex
	for j := 0; j < c.ncpu; j++ {
		wg.Add(1)
		go func(j int) {
			var nodesU [8192]uint64
			last := uint64(steps * (j + 1))
			if j == c.ncpu-1 {
				last += uint64(remain)
			}
		loop:
			for nonce := uint64(steps * j); nonce < last; nonce += 8192 {
				siphash.SiphashPRF8192Seq(&c.sip.V, nonce, 0, &nodesU)
				for i := uint64(0); i < 8192; i++ {
					u0 := nodesU[i] & edgemask
					if es.uxymap[(u0>>zbits)&xymask] {
						v0 := siphash.SiphashPRF(&c.sip.V, ((nonce+i)<<1)|1) & edgemask
						if es.find((u0<<32)|v0, 0, len(es.edge)-1) {
							mutex.Lock()
							answer = append(answer, uint32(nonce+i))
							if len(answer) >= ProofSize {
								mutex.Unlock()
								break loop
							}
							mutex.Unlock()
						}
					}
				}
			}
			wg.Done()
		}(j)
	}
	wg.Wait()
	sort.Slice(answer, func(i, j int) bool {
		return answer[i] < answer[j]
	})
	return answer, true
}

func (c *Cuckoo) buildU() {
	var wg sync.WaitGroup
	// c.matrixs = [8][32][32]uint64, 创建这样一个多维数组，但没有初始化，[[[[] [] [].32.].32.].8.]
	for j := 0; j < c.ncpu; j++ {
		for x := 0; x < nx; x++ { // nx = 2^5
			for y := 0; y < nx; y++ {
				c.matrixs[j][x][y] = c.matrixs[j][x][y][:0]
			}
		}
	}
	//fmt.Printf("buildU() c.matrixs = %v\n",c.matrixs)

	steps := Easiness / c.ncpu        // easiness = 16777216 = 2^24, ncpu = 8, steps = 2097152 = 2^21
	remain := Easiness - steps*c.ncpu // remain = 0
	//TODO, 测试用设置 c.ncpu = 1
	//c.ncpu = 1
	for j := 0; j < c.ncpu; j++ {
		wg.Add(1)
		go func(j int) {
			// 0
			//j = 0
			//last = 2097152
			// 1
			//j = 0
			//last = 16777216
			//2
			//j = 0
			//last = 16777216
			last := uint64((j + 1) * steps)
			//fmt.Printf("***j = %v\nlast = %v\n",j,last)
			if j == c.ncpu-1 {
				last += uint64(remain)
			}
			var nodesU [8192]uint64
			//TODO, 测试用，设置 last = 3
			//last = 3
			for nonce := uint64(steps * j); nonce < last; nonce += 8192 {

				//当 nonce = 0 生成nodesU数组，     其中有8192个随机uint64类型的数字, [9763790239996453423 48331414249040509 ... 6838208080394788122]
				//当 nonce = 8192 生成nodesU数组，  其中有8192个随机uint64类型的数字, [5011153096535506704 14743687886315419173 ... 16895110234403976984]
				//当 nonce = 16384 生成nodesU数组， 其中有8192个随机uint64类型的数字, [14874194351879764632 13369115042760596344 ...13528283603548748680]
				siphash.SiphashPRF8192Seq(&c.sip.V, nonce, 0, &nodesU)
				//fmt.Printf("nonce = %v\n",nonce)
				//fmt.Printf("nodesU = %v\n",nodesU)
				//time.Sleep(1 * time.Second)

				// 根据nodesU = [15097994728306670538 5135416001628188714 ...14861455826150337964]
				// 生成c.matrixs =
				for i := range nodesU {
					//fmt.Printf("i = %v\n",i)
					//fmt.Printf("nodesU[i] = %v\n",nodesU[i])
					u := nodesU[i] & edgemask // 模数运算符，以及整数除法的余数，使得u的值落在[0, 2^24-1]这个区间内
					//fmt.Printf("u = %v\n",u)
					if u == 0 {
						continue
					}
					ux := (u >> (edgebits - xbits)) & xmask
					//fmt.Printf("ux = %v\n",ux)
					uy := (u >> (edgebits - 2*xbits)) & xmask
					//fmt.Printf("uy = %v\n",uy)
					v := ((nonce + uint64(i)) << 32) | u
					//fmt.Printf("nonce = %v\n",nonce)
					//fmt.Printf("v = %v\n",v)
					c.matrixs[j][ux][uy] = append(c.matrixs[j][ux][uy], v)
					//fmt.Printf("c.matrixs = %v\n",c.matrixs)
					//time.Sleep(3 * time.Second)
				}
				//fmt.Printf("c.matrixs = %v\n",c.matrixs)
				//time.Sleep(1 * time.Second)
			}
			wg.Done()
		}(j)
	}
	wg.Wait()
	for j := 0; j < c.ncpu; j++ {
		for x := 0; x < nx; x++ {
			for y := 0; y < nx; y++ {
				c.matrix[x][y] = append(c.matrix[x][y], c.matrixs[j][x][y]...)
			}
		}
	}
}

func (c *Cuckoo) buildV() int {
	var wg sync.WaitGroup
	//TODO, 测试用设置ncpu=1
	c.ncpu = 1

	num := make([]int, c.ncpu)
	steps := nx / c.ncpu // nx = 2^5 = 32
	remain := nx - steps*c.ncpu
	for j := 0; j < c.ncpu; j++ {
		//println("buildV() j = ",j)
		wg.Add(1)
		go func(j int) {
			var nodesV [8192]uint64
			var nonces [8192]uint64
			var us [8192]uint64
			for i := range c.m2tmp[j] {
				c.m2tmp[j][i] = c.m2tmp[j][i][:0]
			}
			last := (j + 1) * steps // last = 32
			//println("buildV() last = ",last)
			if j == c.ncpu-1 {
				last += remain
			}
			//TODO,测试用，设置last = 3
			//last = 3
			for ux := j * steps; ux < last; ux++ {
				mu := c.matrix[ux]
				nsip := 0
				//fmt.Println("buildV() mu = ",mu)
				for _, m := range mu {
					var cnt [nz]byte
					//println("len(m) = ",len(m))
					//println("len(cnt) = ",len(cnt))
					for _, nu := range m {
						cnt[nu&zmask]++
					}
					for _, nu := range m {
						if cnt[nu&zmask] == 1 {
							continue
						}
						num[j]++
						//fmt.Printf("buildV() nu = %v\n",nu)
						//fmt.Printf("buildV() nonces[nsip] = %v\n",nonces[nsip])
						nonces[nsip] = nu >> 32 // nu / 2^32
						//fmt.Printf("buildV() nonces[nsip] = %v\n",nonces[nsip])
						us[nsip] = nu << 32 // nu * 2^32
						//fmt.Printf("buildV() us[nsip] = %v\n",us[nsip])
						//fmt.Printf("buildV() nsip = %v\n",nsip)
						if nsip++; nsip == 8192 {
							nsip = 0
							siphash.SiphashPRF8192(&c.sip.V, &nonces, 1, &nodesV)
							for i, v := range nodesV {
								v &= edgemask // 模数运算符，以及整数除法的余数，使得v的值落在[0, 2^24-1]这个区间内
								vx := (v >> (edgebits - xbits)) & xmask
								c.m2tmp[j][vx] = append(c.m2tmp[j][vx], us[i]|v)
							}
						}
					}
				}
				siphash.SiphashPRF8192(&c.sip.V, &nonces, 1, &nodesV)
				//fmt.Printf("buildV() nonces = %v\n",nonces)
				//fmt.Printf("buildV() nodesV = %v\n",nodesV)
				//time.Sleep(1 * time.Second)
				for i := 0; i < nsip; i++ {
					v := nodesV[i] & edgemask
					//fmt.Printf("v = %v\n",v)
					//time.Sleep(1 * time.Second)
					vx := (v >> (edgebits - xbits)) & xmask
					c.m2tmp[j][vx] = append(c.m2tmp[j][vx], us[i]|v)
				}
				c.matrix[ux], c.m2tmp[j] = c.m2tmp[j], c.matrix[ux]
				for i := range c.m2tmp[j] {
					c.m2tmp[j][i] = c.m2tmp[j][i][:0]
				}
			}
			wg.Done()
		}(j)
	}
	wg.Wait()
	number := 0
	for j := 0; j < c.ncpu; j++ {
		number += num[j]
	}
	return number
}

func (c *Cuckoo) trim(isU bool) (int, int) {
	var wg sync.WaitGroup
	num := make([]int, c.ncpu)
	maxbucket := make([]int, c.ncpu)
	steps := nx / c.ncpu
	remain := nx - steps*c.ncpu
	for j := 0; j < c.ncpu; j++ {
		wg.Add(1)
		go func(j int) {
			last := uint32((j + 1) * steps)
			if j == c.ncpu-1 {
				last += uint32(remain)
			}
			for ux := uint32(j * steps); ux < last; ux++ {
				indexer := c.index(isU, ux)
				for vx := uint32(0); vx < nx; vx++ {
					m := indexer[vx]
					for _, uv := range *m {
						y := (uv >> (edgebits - 2*xbits)) & xmask
						c.m2[j][y] = append(c.m2[j][y], uv)
					}
					*m = (*m)[:0]
				}
				for i, m2y := range c.m2[j] {
					var cnt [nz]byte
					for _, uv := range m2y {
						cnt[uv&zmask]++
					}
					nbucket := 0
					for _, uv := range m2y {
						if cnt[uv&zmask] == 1 {
							continue
						}
						nbucket++
						num[j]++
						vu := uv >> 32
						vux := (vu >> (edgebits - xbits)) & xmask
						ruv := (uv << 32) | vu
						m := indexer[vux]
						*m = append(*m, ruv)
					}
					c.m2[j][i] = c.m2[j][i][:0]
					if maxbucket[j] < nbucket {
						maxbucket[j] = nbucket
					}
				}
			}
			wg.Done()
		}(j)
	}
	wg.Wait()
	number := 0
	maxb := 0
	for j := 0; j < c.ncpu; j++ {
		number += num[j]
		if maxb < maxbucket[j] {
			maxb = maxbucket[j]
		}
	}
	//fmt.Println("number=",number)
	//fmt.Println("maxb=",maxb)
	return number, maxb
}

func (c *Cuckoo) trimrename0(isU bool) int {
	num := 0
	for ux := uint32(0); ux < nx; ux++ {
		indexer := c.index(isU, ux)
		for vx := uint32(0); vx < nx; vx++ {
			m := indexer[vx]
			for _, uv := range *m {
				y := (uv >> (edgebits - 2*xbits)) & xmask
				c.m2[0][y] = append(c.m2[0][y], uv)
			}
			*m = (*m)[:0]
		}
		for i, m2y := range c.m2[0] {
			var nodeid byte
			var cnt [nz]byte
			for _, uv := range m2y {
				cnt[uv&zmask]++
			}
			for _, uv := range m2y {
				cntv := cnt[uv&zmask]
				if cntv == 1 {
					continue
				}
				num++
				var myid byte
				if cntv >= 32 {
					myid = cntv - 32
				} else {
					myid = nodeid
					cnt[uv&zmask] = 32 + nodeid
					nodeid++
				}
				newuv := uv & 0xffffffff
				newuv >>= zbits
				newuv |= (uv & zmask) << (2 * xbits)
				newuv <<= comp0bits
				newuv |= uint64(myid)
				vu := uv >> 32
				allbits := uint(edgebits)
				if isU {
					allbits = 2*xbits + comp0bits
				}
				vux := (vu >> (allbits - xbits)) & xmask
				ruv := (newuv << 32) | vu
				m := indexer[vux]
				*m = append(*m, ruv)
			}
			c.m2[0][i] = c.m2[0][i][:0]
		}
	}
	return num
}

func (c *Cuckoo) trim2(isU bool) int {
	num := 0
	c.m2tmp[0][0] = c.m2tmp[0][0][:0]
	for ux := uint32(0); ux < nx; ux++ {
		var cnt [1 << (xbits + comp0bits)]byte
		indexer := c.index(isU, ux)
		for vx := uint32(0); vx < nx; vx++ {
			m := indexer[vx]
			for _, uv := range *m {
				cnt[uv&ycomp0mask]++
			}
		}
		for vx := uint32(0); vx < nx; vx++ {
			m := indexer[vx]
			for i := len(*m) - 1; i >= 0; i-- {
				uv := (*m)[i]
				if cnt[uv&ycomp0mask] == 1 {
					continue
				}
				num++
				c.m2tmp[0][0] = append(c.m2tmp[0][0], (uv<<32)|(uv>>32))
			}
			*m, c.m2tmp[0][0] = c.m2tmp[0][0], *m
			c.m2tmp[0][0] = c.m2tmp[0][0][:0]
		}
	}
	return num
}

func (c *Cuckoo) trimrename1(isU bool) int {
	num := 0
	for ux := uint32(0); ux < nx; ux++ {
		indexer := c.index(isU, ux)
		for vx := uint32(0); vx < nx; vx++ {
			m := indexer[vx]
			for _, uv := range *m {
				y := (uv >> comp0bits) & xmask
				c.m2[0][y] = append(c.m2[0][y], uv)
			}
			*m = (*m)[:0]
		}
		for i, m2y := range c.m2[0] {
			var nodeid byte
			var cnt [nz]byte
			for _, uv := range m2y {
				cnt[uv&comp0mask]++
			}
			for _, uv := range m2y {
				cntv := cnt[uv&comp0mask]
				if cntv == 1 {
					continue
				}
				num++
				var myid byte
				if cntv >= 32 {
					myid = cntv - 32
				} else {
					myid = nodeid
					cnt[uv&comp0mask] = 32 + nodeid
					nodeid++
				}
				newuv := uv & 0xffffffff
				newuv >>= comp0bits
				newuv <<= comp1bits
				newuv |= uint64(myid)
				vu := uv >> 32
				nbits := uint(comp0bits)
				if isU {
					nbits = comp1bits
				}
				vux := (vu >> (nbits + xbits)) & xmask
				ruv := (newuv << 32) | vu
				m := indexer[vux]
				*m = append(*m, ruv)
			}
			c.m2[0][i] = c.m2[0][i][:0]
		}
	}
	return num
}

func newedges() *edges {
	return &edges{
		edge:   make([]uint64, 0, ProofSize),
		uxymap: make([]bool, 1<<(2*xbits)),
	}
}

func (e *edges) add(u, v uint32) {
	u >>= 1
	uz := u >> (2*xbits + comp1bits)
	uxy := (u >> comp1bits) & xymask
	ru := (uxy << zbits) | uz
	e.uxymap[uxy] = true
	v >>= 1
	vz := v >> (2*xbits + comp1bits)
	vxy := (v >> comp1bits) & xymask
	rv := (vxy << zbits) | vz
	e.edge = append(e.edge, (uint64(ru)<<32)|uint64(rv))
}

func (e *edges) find(uv uint64, min, max int) bool {
	if max < min {
		return false
	}
	mid := (min + max) / 2
	if e.edge[mid] > uv {
		return e.find(uv, min, mid-1)
	}
	if e.edge[mid] < uv {
		return e.find(uv, mid+1, max)
	}
	return true
}

func (c *Cuckoo) index(isU bool, x uint32) [nx]*bucket {
	var indexer [nx]*bucket
	if isU {
		for i := 0; i < nx; i++ {
			indexer[i] = &c.matrix[x][i]
		}
		return indexer
	}
	for i := 0; i < nx; i++ {
		indexer[i] = &c.matrix[i][x]
	}
	return indexer
}

func MakePowHeaderBytes(header *types.BlockHeader) []byte {
	var headerBytes []byte

	// Version
	versionBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(versionBytes, uint32(header.Version))
	headerBytes = append(headerBytes, versionBytes...)

	// Previous block hash
	var prevBlockHashArr [32]byte
	var prevBlockHash []byte
	prevBlockHashArr = [32]byte(header.ParentRoot)
	prevBlockHash = prevBlockHashArr[:]
	headerBytes = append(headerBytes, prevBlockHash...)

	// Merkle root hash
	var merkleRootArr [32]byte
	var merkleRoot []byte
	merkleRootArr = [32]byte(header.TxRoot)
	merkleRoot = merkleRootArr[:]
	headerBytes = append(headerBytes, merkleRoot...)

	// Time
	timeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(timeBytes, uint32(header.Timestamp.Unix()))
	headerBytes = append(headerBytes, timeBytes...)

	// Target bits
	bitsBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bitsBytes, header.Difficulty)
	headerBytes = append(headerBytes, bitsBytes...)

	// Nonce
	nonceBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonceBytes, header.Nonce)
	headerBytes = append(headerBytes, nonceBytes...)

	return headerBytes
}

func ComputeBlockHeaderHash(header []byte) []byte {
	hash := computeBTCHash(header)
	return reverseBytes(hash[:])
}

func computeBTCHash(data []byte) []byte {
	h1 := sha256.Sum256(data)
	h2 := sha256.Sum256(h1[:])
	return h2[:]
}

func reverseBytes(bytes []byte) []byte {
	for i, j := 0, len(bytes)-1; i < j; i, j = i+1, j-1 {
		bytes[i], bytes[j] = bytes[j], bytes[i]
	}
	return bytes
}
