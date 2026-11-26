package processor

import "sync"

// ProcessFiles возвращает по результату на каждый переданный путь.
// Получатель читает канал до закрытия.
func ProcessFiles(paths []string, workers int) <-chan Result {
	results, _ := ProcessFilesWithCommon(paths, workers)
	return results
}

func ProcessFilesWithCommon(paths []string, workers int) (<-chan Result, *Common) {
	if workers < 1 {
		workers = 1
	}

	common := Common{0, 0}
	results := make(chan Result)
	files := make(chan string)

	go func() {
		for _, path := range paths {
			files <- path
		}
		close(files)
	}()

	waitGroup := new(sync.WaitGroup)
	mutex := new(sync.Mutex)

	for i := 0; i < workers; i++ {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()
			for file := range files {
				part := analyzeFile(file)
				results <- part

				mutex.Lock()
				common.Lines += part.Lines
				common.Words += part.Words
				mutex.Unlock()

			}
		}()
	}

	go func() {
		waitGroup.Wait()
		close(results)
	}()

	return results, &common
}
