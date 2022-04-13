package main

type Configuration struct {
	Selenium SeleniumConfiguration
}

type SeleniumConfiguration struct {
	Url             string
	PageLoadTimeout int
}
