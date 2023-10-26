package utils

import (
	"fmt"
)

type Progresser struct {
	name   string
	wrote  int64
	size   int64
	pCh    chan<- string
	prc    int64
	prcCnt int64
}

func NewProgresser(name string, size int64, pCh chan<- string) *Progresser {
	return &Progresser{
		name: name,
		size: size,
		pCh:  pCh,
		prc:  size / 100 * 5,
	}
}

func (p *Progresser) Write(data []byte) (int, error) {
	p.wrote += int64(len(data))
	p.prcCnt += int64(len(data))

	if p.prcCnt > p.prc || p.wrote == p.size {
		p.pCh <- fmt.Sprintf("%s: %f%%/100%%", p.name, float64(p.wrote)/float64(p.size)*100.0)
		p.prcCnt = 0
	}

	return len(data), nil
}
