package port

import "sync"

type Map map[string]interface{}

var mapPool = sync.Pool{
	New: func() interface{} {
		return make(Map)
	},
}

func MakeMap(args ...interface{}) Map {
	if len(args)%2 != 0 {
		panic("uneven number of arguments to port.MakeMap")
	}

	m := mapPool.Get().(Map)
	for i := 0; i+1 < len(args); i += 2 {
		m[args[i].(string)] = args[i+1]
	}
	return m
}

func RecycleMap(m Map) {
	if m != nil {
		for k, v := range m {
			delete(m, k)
			if _, ok := v.(*Port); ok {
				portPool.Put(v)
			}
		}
		mapPool.Put(m)
	}
}
