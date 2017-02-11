package parallel

import "sync"

// ProduceFunc is a function that generates something and puts them in products
type ProduceFunc func(products chan<- interface{}, done <-chan struct{})

// ConsumeFunc is a function that consumes products (processes data)
type ConsumeFunc func(product interface{}, done <-chan struct{})

// Runner contains all parameters for a runner
type Runner struct {
	Produce      ProduceFunc
	Consume      ConsumeFunc
	NumProducers int
	NumConsumers int
	Logger       simpleLogger
	ConsumeAll   bool
	waitGroup    sync.WaitGroup
	done         chan struct{}
}

// New returns a new Runner
func New() (r *Runner) {
	r = &Runner{
		NumProducers: 1,
		NumConsumers: 5,
		Logger:       initLogger(),
		ConsumeAll:   true,
	}
	return
}

// Run starts the Runner
func (r *Runner) Run() {
	if r.Produce == nil {
		r.Logger.Fatalln("Please specify a Produce function")
	}

	if r.Consume == nil {
		r.Logger.Fatalln("Please specify a Consume function")
	}

	r.done = make(chan struct{})
	defer r.Stop()

	r.Logger.Println("Runner started ...")
	products := r.producer()
	r.consumer(products)
	r.waitGroup.Wait()
	r.Logger.Println("Runner finished.")
}

// Stop closes done channel to stop all producers and consumers
func (r *Runner) Stop() {
	select {
	case <-r.done:
	default:
		close(r.done)
		r.Logger.Println("Runner stopped.")
	}
}

func (r *Runner) producer() <-chan interface{} {
	products := make(chan interface{}, r.NumConsumers)
	var wg sync.WaitGroup
	wg.Add(r.NumProducers)

	go func() {
		wg.Wait()
		close(products)
	}()

	r.waitGroup.Add(r.NumProducers)
	for i := 0; i < r.NumProducers; i++ {
		go func(id int) {
			r.Logger.Printf("Producer %d started.", id)
			r.Produce(products, r.done)
			r.waitGroup.Done()
			wg.Done()
			r.Logger.Printf("Producer %d finished.", id)
		}(i)
	}
	return products
}

func (r *Runner) consumer(products <-chan interface{}) {
	r.waitGroup.Add(r.NumConsumers)
	for i := 0; i < r.NumConsumers; i++ {
		go func(id int) {
			r.Logger.Printf("Consumer %d started.", id)
			for product := range products {
				if r.ConsumeAll {
					r.Consume(product, r.done)
				} else {
					select {
					case <-r.done:
						break
					default:
						r.Consume(product, r.done)
					}
				}
			}
			r.waitGroup.Done()
			r.Logger.Printf("Consumer %d finished.", id)
		}(i)
	}
}
