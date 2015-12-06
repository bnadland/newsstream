package main

func main() {
	config := NewConfig()

	newsstream, err := NewNewsstream(config)
	if err != nil {
		log.Error("Unable to init app: ", err)
		return
	}

	newsstream.Run()
}
