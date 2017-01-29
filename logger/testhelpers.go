package logger

// The following are to facilitate unit tests. "_test.go" files are not, by default, exported outside
// of their package, so this just makes these basic structures available to other test packages

// TestingWriter is a simple object that implements io.Writer for use in testing environments
type TestingWriter struct {
	// It might seem a little silly that this is a pointer to a string slice rather than just a value
	// but I assure you there's a good reason. Unfortunately, to implement io.Writer, the Write
	// method *must* be pass by value. This obviously means that editing "previous" as a value
	// in the Write method would be entirely ephemeral, as we're essentially setting it on a copy
	// of our "actual" TestingWriter. If, however, we make previous a pointer then our copy
	// will be pointing to the same memory address as our "actual" object, allowing us to set
	// the value at that memory location so that the value can be retrieved.
	//
	// You might also be thinking that all of this is silly because, under the covers, isn't an array
	// just a pointer to an element anyway? Why do we need a pointer to a pointer? That is totally true,
	// even in Go an array is simply a pointer to the first element and a counter of how many elements
	// there are. The issue is that here we're not using an array, we're using a *slice*, which means
	// there is an undefined number of elements. Now this isn't implemented like a linked-list or anything,
	// where the first element's memory address can remain relatively static (those wouldn't be fast enough).
	// Instead, to append an element, you call "slice = append(slice, element)". Behind the scenes, "append"
	// will try to append the element to the existing underlying array if there's space, if there's not then
	// it will have to allocate a larger (mostly likely double in size) array and copy everything over, returning
	// the new memory address (this is also, roughly, how C++'s vectors work, by the way). We can't depend on the
	// array at initialization being the same array after we're done appending to it. So, in this case, we really
	// do want a pointer to a pointer.
	//
	// This is all fairly 101, to be honest, but it's worth pointing out to illustrate that while Go
	// will happily manage your memory allocations and deallocations, it's still necessary to maintain
	// a working understanding of traditional memory principles or else you can easily shoot yourself in
	// the foot.
	previous *[]string
}

// CreateTestingWriter creates an instance of a TestingWriter.
//
// While no doubt rejecting many of it's traditional concepts, I still maintain that Go is object
// oriented. You're expected to define reusable structures which can have instance level functions
// associated with it. As such, I find it frustrating that it chooses to simply ignore the concept
// of initialization. The Go justification that your structs should just work when not initialized
// is just naive.
// Take for instance the simple case of TestingWriter which, for reasons previously stated, is composed
// of a pointer to a string. Now if you were to create an instance using only "TestingWriter{}" you'll
// quickly notice, on the first call to the Write method, you'll encounter a nasty "invalid memory address"
// runtime error, because our pointer is, obviously, not initialized and thus not pointing to a valid
// address. Following Go's advice, maybe we should add a check to our Write method to see if our "previous"
// pointer is null and, if so, initialize it. However this wouldn't work in our case, we'd be initializing the
// pointer for the passed in TestingWriter *value* only.
//
// Well in that case simply create instances with TestingWriter{new(string)} instead. That's a valid approach in
// that it works but now we have a completely unenforced (in code, at least) requirement for how TestingWriter
// *has* to be initialized. And all it takes is one mistake where we forget and write TestingWriter{} to blow
// up our code. A small enough risk in this test scenerio but this is a recurrent problem for *all* object
// definitions in Go and you can extrapolate how this might allow us to accidentally put a ticking time bomb somewhere
// nested in production code. In this scenerio, what we *want* (and I'd argue, need) is proper RAII. However
// it doesn't exist in Go, so the best we can do is define unofficial "constructors" like this and hope that
// people use it.
func CreateTestingWriter() TestingWriter {
	return TestingWriter{new([]string)}
}

func (writer TestingWriter) Write(p []byte) (n int, err error) {
	*writer.previous = append(*writer.previous, string(p))
	return len(p), nil
}

// Last returns the last written log statement
func (writer TestingWriter) Last() string {
	// I don't want to be *that* guy, but why can't we all just freaking agree that python slice notations
	// would make all of our lives easier?
	length := len(*writer.previous)
	if length == 0 {
		// I'm not happy with this. This return should denote a totally invalid entry, of which an empty string
		// is not (there's nothing stopping you from logging one). It's not possible to return nil here (the compiler
		// won't let us), so the only other alternative would be to have multiple returns, the second of which is an
		// error. The problem with that, in this context, is that Last should be something you can call in a single
		// value context (in an assert, for instance), and forcing the caller to handle multiple values for each call
		// would quickly become a nuisance. If this were production code I like to think I would take a firmer stand,
		// but writing tests is painful enough, so let's just return a single value.
		return ""
	}

	return (*writer.previous)[length-1]
}

// DummyLogger provides a simple collection of writers for testing using the logger
type DummyLogger struct {
	Debug   TestingWriter
	Info    TestingWriter
	Warning TestingWriter
	Error   TestingWriter
}

// CreateDummyLogger creates an instance of a DummyLogger and initializes it's writers
// with the application loggers
func CreateDummyLogger() DummyLogger {
	logger := DummyLogger{
		Debug:   CreateTestingWriter(),
		Info:    CreateTestingWriter(),
		Warning: CreateTestingWriter(),
		Error:   CreateTestingWriter(),
	}
	InitLogging(logger.Debug, logger.Info, logger.Warning, logger.Error, 0)
	return logger
}
