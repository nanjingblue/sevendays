package singleflight

import "sync"

type Call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex
	m  map[string]*Call
}

/*
Do 方法，接收 2 个参数，第一个参数是 key，第二个参数是一个函数 fn。
Do 的作用就是，针对相同的 key，无论 Do 被调用多少次，函数 fn 都只会被调用一次，
等待 fn 调用结束了，返回返回值或错误。
*/
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*Call)
	}

	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		/*
			wg.Add(1) 锁加1。
			wg.Wait() 阻塞，直到锁被释放。
			wg.Done() 锁减1
		*/
		c.wg.Wait()
		return c.val, c.err
	}

	c := &Call{}
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
	return c.val, c.err
}
