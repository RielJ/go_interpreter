package object

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

// Environment
type Environment struct {
	store map[string]Object
	outer *Environment
}

// Get
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		// If the object is not found in the current environment, check the outer environment
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}
