package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

type arrayFlags []string

func (af *arrayFlags) String() string {
	return "Some text"
}

func (af *arrayFlags) Set(v string) error {
	*af = append(*af, v)
	return nil
}

func main() {
	var companies arrayFlags
	var titles arrayFlags
	var locations arrayFlags

	login := flag.String("login", "", "Login for LinkedIn account")
	password := flag.String("password", "", "Password for LinkedIn account")
	removeDisabled := flag.Bool("remove-disabled", false, "Remove disabled elements")

	flag.Var(&companies, "company", "Company to remove from result")
	flag.Var(&titles, "title", "Job title to remove from result")
	flag.Var(&locations, "location", "Job location to remove from result")

	flag.Parse()

	// Run Chrome browser
	service, err := selenium.NewChromeDriverService("./chromedriver", 4444)
	if err != nil {
		panic(err)
	}
	defer service.Stop()

	caps := selenium.Capabilities{}
	caps.AddChrome(chrome.Capabilities{Args: []string{
		"window-size=1920x1080",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"disable-gpu",
		// "--headless",  // comment out this line to see the browser
	}})

	driver, err := selenium.NewRemote(caps, "")
	if err != nil {
		panic(err)
	}

	driver.Get("https://linkedin.com")
	driver.SetImplicitWaitTimeout(1 * time.Second)

	el, err := driver.FindElement(selenium.ByID, "session_key")
	if err != nil {
		panic("Failed to find session key")
	}
	el.SendKeys(*login)

	el, err = driver.FindElement(selenium.ByID, "session_password")
	if err != nil {
		panic("Failed to find session password")
	}
	el.SendKeys(*password)

	el, err = driver.FindElement(selenium.ByXPATH, "//*[@id=\"main-content\"]/section[1]/div/div/form/div[2]/button")
	if err != nil {
		panic("Failed to find sign in button")
	}
	el.Click()

	el, err = driver.FindElement(selenium.ByXPATH, "//*[@id=\"global-nav-typeahead\"]/input")
	if err != nil {
		panic("Failed to find search element")
	}

	el.SendKeys("golang\n")

	driver.SetImplicitWaitTimeout(5 * time.Second)

	el, err = driver.FindElement(selenium.ByXPATH, "//*[@id=\"search-reusables__filters-bar\"]/ul/li[1]/button")
	if err != nil {
		panic("Failed to find Jobs button")
	}

	el.Click()

	driver.SetImplicitWaitTimeout(3 * time.Second)

	Loop(driver, companies, titles, locations, *removeDisabled)
}

func Loop(driver selenium.WebDriver, companies, titles, locations arrayFlags, removeDisabled bool) {
	inputScan := make(chan string)
	go func() {
		defer close(inputScan)
		for {
			input := bufio.NewScanner(os.Stdin)
			input.Scan()

			if len(input.Text()) == 0 {
				continue
			}

			inputScan <- input.Text()
			if input.Text()[0] == 'q' {
				break
			}
		}
	}()

	for {
		select {
		case input := <-inputScan:
			if input[0] == 'q' {
				return
			} else if input[0] == 'l' {
				locations = append(locations, input[1:])
			} else if input[0] == 'c' {
				companies = append(companies, input[1:])
			} else if input[0] == 't' {
				titles = append(titles, input[1:])
			}
		case <-time.After(200 * time.Millisecond):
			els, _ := driver.FindElements(selenium.ByClassName, "jobs-search-results__list-item")

			for _, v := range els {
				id, _ := v.GetAttribute("id")
				if len(id) == 0 {
					continue
				}

				if removeDisabled {
					_, err := v.FindElement(selenium.ByClassName, "job-card-list--is-dismissed")
					if err == nil {
						RemoveElement(id, driver)
						fmt.Println("Removing disabled element")
					}
				}

				el, err := v.FindElement(selenium.ByClassName, "job-card-container__metadata-item")
				if err == nil {
					location, _ := el.Text()
					for _, l := range locations {
						if strings.Contains(location, l) {
							HideElement(id, driver)
							RemoveElement(id, driver)
							fmt.Println("Removing location", location)
							continue
						}
					}
				}

				el, err = v.FindElement(selenium.ByClassName, "job-card-container__primary-description")
				if err == nil {
					company, _ := el.Text()
					for _, c := range locations {
						if strings.Contains(company, c) {
							fmt.Println("Removing company", company)
							HideElement(id, driver)
							RemoveElement(id, driver)
							continue
						}
					}
				}

				el, err = v.FindElement(selenium.ByClassName, "job-card-list__title")
				if err == nil {
					title, _ := el.Text()
					for _, c := range titles {
						if strings.Contains(title, c) {
							fmt.Println("Removing title", title)
							HideElement(id, driver)
							RemoveElement(id, driver)
							continue
						}
					}
				}
			}
		}
	}
}

func HideElement(id string, driver selenium.WebDriver) error {
	el, err := driver.FindElement(selenium.ByID, id)
	if err != nil {
		return err
	}

	el, err = el.FindElement(selenium.ByClassName, "job-card-list__dismiss")
	if err != nil {
		return err
	}

	el.FindElement(selenium.ByTagName, "button")
	if err != nil {
		return err
	}

	el.Click()
	return nil
}

func RemoveElement(id string, driver selenium.WebDriver) error {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "return document.getElementById('%s').remove();", id)

	_, err := driver.ExecuteScript(buf.String(), nil)
	if err != nil {
		fmt.Println("Failed to execute", buf.String())
		return err
	}

	return nil
}
