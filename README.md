# linkedin-filter
Advance filter for LinkedIn

Allows to remove company/location/job title from job search result

# Usage
Download chrome driver to the main.go directory

https://chromedriver.chromium.org/downloads

```
go run ./... --login <login> --password <password> --company abc --company def --location <location1> --location <location2> --title <title1> --title <title2>
```

Also supporting runtime command l = location, t = title, c = company and q for quiting