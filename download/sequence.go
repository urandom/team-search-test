package download

import (
	"sort"
	"time"
)

type feedback struct {
	id    int
	retry bool
}

func sequence(problemFeedback <-chan feedback) <-chan int {
	ids := make(chan int)

	go func() {
		repeaters := map[int]int{}

		maxErrors := cap(problemFeedback)
		if maxErrors < 100 {
			maxErrors = 100
		}

		maxRepeat := 10
		errIds := make([]int, 0, maxErrors)

		i := 0
		for {
			select {
			case p := <-problemFeedback:
				if p.retry {
					c := repeaters[p.id]

					if c < maxRepeat {
						c++
						repeaters[p.id] = c

						// A network error will likely manifest again unless we
						// give it some time to breathe.
						time.Sleep(50 * time.Millisecond)
						ids <- p.id
					}
				} else {
					// Since there is no sure fire way of determining whether
					// we've obtained all available teams, we'll have to make
					// an educated guess. We'll collect all problematic ids
					// that are not due to some network error into a pool. Once
					// it is filled, we'll sort and determine if the ids are
					// more or less continuous. If there are gaps, it is most
					// likely due to certain teams not existing anymore.
					// Otherwise, we assume that there are no more teams and
					// close the id generator.
					if len(errIds) < cap(errIds) {
						errIds = append(errIds, p.id)
					} else {
						continuous := false

						sort.Ints(errIds)
						// Give some leeway due to the async order of processing
						if errIds[len(errIds)-1]-errIds[0] < maxErrors+cap(problemFeedback) {
							continuous = true
						}

						if continuous {
							close(ids)
							return
						} else {
							errIds = errIds[:0]
						}
					}
				}
			case ids <- i:
				i++
			}
		}
	}()

	return ids
}
