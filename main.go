// file: main.go

// Test URLs:
// http://localhost:8080/?xmin=39117&xmax=39137&ymin=48460&ymax=48473&zoom=17&scale=256
// http://localhost:8080/?tileurl=http://tile.openstreetmap.org/17/39137/48460.png

// Original code (Perl):
// http://openstreetmap.gryph.de/bigmap.cgi?xmin=39117&xmax=39144&ymin=48463&ymax=48484&zoom=17&scale=256&baseurl=http%3A%2F%2Ftile.openstreetmap.org%2F%21z%2F%21x%2F%21y.png
// http://openstreetmap.gryph.de/bigmap.txt

package main

import (
	"fmt"
	"github.com/kataras/iris"
	"github.com/thumbline-forks/gosm"
	"html"
	"html/template"
	"regexp"
	"strconv"
)

const RelativePath = "/"

func main() {
	app := iris.New()

	// Load all templates from the "./views" folder
	// where extension is ".html" and parse them
	// using the standard `html/template` package.
	app.RegisterView(iris.HTML("./views", ".html"))

	app.Get(RelativePath, func(ctx iris.Context) {
		xMin, _ := ctx.URLParamInt64("xmin")
		yMin, _ := ctx.URLParamInt64("ymin")
		xMax, _ := ctx.URLParamInt64("xmax")
		yMax, _ := ctx.URLParamInt64("ymax")
		zoom, _ := ctx.URLParamInt64("zoom")
		scale, _ := ctx.URLParamInt64("scale")
		baseURL := ctx.URLParam("baseurl")
		tileURL := ctx.URLParam("tileurl")

		if tileURL != "" {
			tileURLPngRegex := regexp.MustCompile("^(.*?)/(\\d+)/(\\d+)/(\\d+)\\.png$")
			if tileURLPngRegex.MatchString(tileURL) {
				matches := tileURLPngRegex.FindStringSubmatch(tileURL)
				if len(matches) == 5 {
					/* This will return []matches like:
					[0]="http://tile.openstreetmap.org/17/39118/48460.png"
					[1]="http://tile.openstreetmap.org"
					[2]="17"
					[3]="39118"
					[4]="48460"
					*/
					zoom, _ = strconv.ParseInt(matches[2], 10, 64)
					xMin, _ = strconv.ParseInt(matches[3], 10, 64)
					if xMin < 1 {
						xMin = 1
					}
					xMin--
					xMax = xMin + 2
					yMin, _ = strconv.ParseInt(matches[4], 10, 64)
					if yMin < 1 {
						yMin = 1
					}
					yMin--
					yMax = yMin + 2

					// TODO Stripping relief & trails from the URL
					if baseURL == "" {
						// Original Perl code had this:
						// $baseurl =~ s/\/(relief|trails)//g;
						// I'm just building baseURL from the image name:
						baseURL = matches[1] + "/!z/!x/!y.png"
					}
				}
			}
		}

		if scale == 0 {
			scale = 256
		}
		if baseURL == "" {
			baseURL = "http://tile.openstreetmap.org/!z/!x/!y.png"
		}

		tilesHTML := ""
		splitRegexp := regexp.MustCompile("\\|") // back-slashed "|" (vertical line)

		for y := yMin; y <= yMax; y++ {
			tilesHTML += fmt.Sprintf("\n<!--%d-->\n", y)
			for x := xMin; x <= xMax; x++ {
				urls := splitRegexp.Split(baseURL, -1)
				xp := scale * (x - xMin)
				yp := scale * (y - yMin)
				for _ /*idx*/, overlayURL := range urls {
					tilesHTML += fmt.Sprintf(`<div style="position:absolute; left:%d; top:%d; width: %d; height: %d">`, xp, yp, scale, scale)
					tilesHTML += img(overlayURL, zoom, x, y, scale)
					tilesHTML += `</div>`
				}
			}

		}

		ctx.ViewData("tiles", template.HTML(tilesHTML))
		ctx.ViewData("control", template.HTML(getControlHTML(baseURL, xMin, xMax, yMin, yMax, zoom, scale)))

		ctx.View("bigmap.html")
	})

	// Start the server using a network address.
	app.Run(iris.Addr(":8080"))
}

// img returns an img tag with the url template containing !z, !x, !y, substituted with corresponding parameters
func img(url string, zoom int64, x int64, y int64, scale int64) string {
	url = replaceAllRegexWithInt64(url, "!z", zoom)
	url = replaceAllRegexWithInt64(url, "!x", x)
	url = replaceAllRegexWithInt64(url, "!y", y)
	return fmt.Sprintf(`<img src="%s" width="%d" height="%d" onclick="getElementById('control').style.display='block';">`,
		url, scale, scale)
}

// replaceAllRegexWithInt64 returns an incomingString with all cases of regexpExpression replaced with the newValue
func replaceAllRegexWithInt64(incomingString string, regexpExpression string, newValue int64) string {
	return string(regexp.
		MustCompile(regexpExpression).
		ReplaceAllString(incomingString, strconv.FormatInt(newValue, 10)))
}

func getControlHTML(baseURL string, xMin, xMax, yMin, yMax, zoom, scale int64) string {
	var html string
	html += `<div id="control" style="align:center;position:fixed;top:50px;margin-left:50px;margin-right:50px;padding:10px;background:#ffffff;opacity:0.8;border:solid 1px;border-color:green;">`
	xTilesCount := xMax - xMin + 1
	yTilesCount := yMax - yMin + 1
	xPixels := xTilesCount * 256
	yPixels := yTilesCount * 256
	asp := "1:1"
	if xPixels > yPixels {
		asp = fmt.Sprintf("%.1f:1", float64(xPixels)/float64(yPixels))
	} else if xPixels < yPixels {
		asp = fmt.Sprintf("1:%.1f", float64(yPixels)/float64(xPixels))
	}
	html += fmt.Sprintf("Map is %dx%d tiles (%dx%d px) at zoom %d, aspect %s<br>",
		xTilesCount, yTilesCount, xPixels, yPixels, zoom, asp)
	p1 := gosm.NewTileWithXY(xMin, yMin, zoom)
	p2 := gosm.NewTileWithXY(xMax+1, yMax+1, zoom)
	html += fmt.Sprintf("Tile: %v", p1) + "<br/>"
	html += fmt.Sprintf("Tile: %v", p2) + "<br/>"
	if zoom > 7 {
		html += fmt.Sprintf("Bbox: (%10.6f, %10.6f) - (%10.6f, %10.6f)", p1.Long, p2.Lat, p2.Long, p1.Lat)
	} else {
		html += fmt.Sprintf("Bbox: (%7.2f, %7.2f) - (%7.2f, %7.2f)", p1.Long, p2.Lat, p2.Long, p1.Lat)
	}
	html += `<table cellspacing="0" cellpadding="2"><tr>`
	html += td(baseURL, "tl", "right", xMin-1, xMax, yMin-1, yMax, zoom, scale)
	html += td(baseURL, "top", "center", xMin, xMax, yMin-1, yMax, zoom, scale)
	html += td(baseURL, "tr", "left", xMin, xMax+1, yMin-1, yMax, zoom, scale)
	html += `<td>&nbsp;</td>`
	html += td(baseURL, "ul", "right", xMin-1, xMax-1, yMin-1, yMax-1, zoom, scale)
	html += td(baseURL, "up", "center", xMin, xMax, yMin-1, yMax-1, zoom, scale)
	html += td(baseURL, "ur", "left", xMin+1, xMax+1, yMin-1, yMax-1, zoom, scale)
	html += `<td>&nbsp;</td>`
	html += td(baseURL, "tl", "right", xMin+1, xMax, yMin+1, yMax, zoom, scale)
	html += td(baseURL, "top", "center", xMin, xMax, yMin+1, yMax, zoom, scale)
	html += td(baseURL, "tr", "left", xMin, xMax-1, yMin+1, yMax, zoom, scale)

	html += `</tr><tr>`
	html += td(baseURL, "left", "right", xMin-1, xMax, yMin, yMax, zoom, scale)
	html += `<td align='center' bgcolor='#aaaaaa'><b>EXPAND</b></td>`
	html += td(baseURL, "right", "left", xMin, xMax+1, yMin, yMax, zoom, scale)
	html += `<td>&nbsp;</td>`
	html += td(baseURL, "left", "right", xMin-1, xMax-1, yMin, yMax, zoom, scale)
	html += `<td align='center' bgcolor='#aaaaaa'><b>SHIFT</b></td>`
	html += td(baseURL, "right", "left", xMin+1, xMax+1, yMin, yMax, zoom, scale)
	html += `<td>&nbsp;</td>`
	html += td(baseURL, "left", "right", xMin+1, xMax, yMin, yMax, zoom, scale)
	html += `<td align='center' bgcolor='#aaaaaa'><b>SHRINK</b></td>`
	html += td(baseURL, "right", "left", xMin, xMax-1, yMin, yMax, zoom, scale)

	html += `</tr><tr>`
	html += td(baseURL, "bl", "right", xMin-1, xMax, yMin, yMax+1, zoom, scale)
	html += td(baseURL, "bottom", "center", xMin, xMax, yMin, yMax+1, zoom, scale)
	html += td(baseURL, "br", "left", xMin, xMax+1, yMin, yMax+1, zoom, scale)
	html += `<td>&nbsp;</td>`
	html += td(baseURL, "dl", "right", xMin-1, xMax-1, yMin+1, yMax+1, zoom, scale)
	html += td(baseURL, "down", "center", xMin, xMax, yMin+1, yMax+1, zoom, scale)
	html += td(baseURL, "dr", "left", xMin+1, xMax+1, yMin+1, yMax+1, zoom, scale)
	html += `<td>&nbsp;</td>`
	html += td(baseURL, "bl", "right", xMin+1, xMax, yMin, yMax-1, zoom, scale)
	html += td(baseURL, "bottom", "center", xMin, xMax, yMin, yMax-1, zoom, scale)
	html += td(baseURL, "br", "left", xMin, xMax-1, yMin, yMax-1, zoom, scale)

	html += `</tr><tr><td></td></tr>`
	html += `<tr><td colspan='11'><table bgcolor='#aaaaaa' width='100%' border='0' cellpadding='0' cellspacing='0'><tr>`
	html += `<td>&nbsp;</td>`
	html += td(baseURL, "in/double size", "left", xMin*2, xMax*2+1, yMin*2, yMax*2+1, zoom+1, scale)
	html += `<td>&nbsp;</td>`
	html += td(baseURL, "in/keep size", "left", xMin*2+(xMax-xMin)/2, xMax*2-(xMax-xMin)/2, yMin*2+(yMax-yMin)/2, yMax*2-(yMax-yMin)/2, zoom+1, scale)
	html += `<td>&nbsp;</td>`
	html += `<td bgcolor='#aaaaaa'><b>ZOOM</b></td>`
	html += `<td>&nbsp;</td>`
	html += td(baseURL, "out/keep size", "left", xMin/2-(xMax-xMin)/4, xMax/2+(xMax-xMin)/4, yMin/2-(yMax-yMin)/4, yMax/2+(yMax-yMin)/4, zoom-1, scale)
	html += `<td>&nbsp;</td>`
	html += td(baseURL, "out/halve size", "left", xMin/2, xMax/2, yMin/2, yMax/2, zoom-1, scale)
	html += `</tr></table></td></tr><tr><td></td></tr>`
	html += `<tr><td colspan='11'><table bgcolor='#aaaaaa' width='100%' border='0' cellpadding='0' cellspacing='0'><tr>`
	html += `<td>&nbsp;</td>`
	html += td(baseURL, "Permalink", "left", xMin, xMax, yMin, yMax, zoom, scale)
	html += `<td>&nbsp;</td>`
	html += td(baseURL, "100%", "left", xMin, xMax, yMin, yMax, zoom, 256)
	html += td(baseURL, "50%", "left", xMin, xMax, yMin, yMax, zoom, 128)
	html += td(baseURL, "25%", "left", xMin, xMax, yMin, yMax, zoom, 64)
	html += `<td align='right'><a href="#" onclick="getElementById('control').style.display='none';">hide this (click map to show again)</a></td>`
	html += `</tr></table>`
	/*
		my $fm = td("Form", "left", xMin,xMax,yMin,yMax,zoom);
		$fm =~ s/\?/?form=1&/;
		print $fm;
		print "<td>&nbsp;</td>";
		my $pl = td("Perl", "left", xMin,xMax,yMin,yMax,zoom);
		$pl =~ s/\?/?perl=1&/;
		print $pl;
	*/
	return html + "</div>"
}

func td(baseurl, what, align string, xmi, xma, ymi, yma, zm, scl int64) string {
	return fmt.Sprintf(`<td bgcolor="#aaaaaa" align="%s"><a href="?xmin=%d&xmax=%d&ymin=%d&ymax=%d&zoom=%d&scale=%d&baseurl=%s">%s</a></td>`,
		align, xmi, xma, ymi, yma, zm, scl, html.EscapeString(baseurl), what)
}
