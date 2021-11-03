function spline(chart, d, valueIndex, name) {
    const s = chart.spline(d.mapAs({x: 0, value: valueIndex}));
    s.name(name)
    s.markers().zIndex(100)
    s.hovered().markers().enabled(true).type('circle').size(4)
}

function splineChart(title, xTitle, yTitle) {
    const c = anychart.line();
    c.yAxis().labels().format('{%Value}')
    c.animation(true)
    c.crosshair().enabled(true).yLabel({enabled: false}).yStroke(null).xStroke('#cecece').zIndex(99)
    c.yAxis().title(yTitle).labels({padding: [5, 5, 0, 5]})
    c.xAxis().title(xTitle)
    c.title(title)
    return c
}

function spliceDraw(c, divId) {
    c.legend().enabled(true).fontSize(13).padding([0, 0, 20, 0])
    c.container(divId)
    c.draw()
}

function findColumnIndex(columns, column) {
    for (let i = 0; i < columns.length; i++) {
        if (columns[i] === column) {
            return i
        }
    }

    return -1
}

function drawChart() {
    let showHeaders = [];
    for (let i = 0; i < headers.length; i++) {
        if (headers[i].endsWith("-MEM") || headers[i].endsWith("-RES")) {
            showHeaders.push(headers[i])
        }
    }

    const c = splineChart('Process RES', 'Time', 'Memory Usage(K/M/G)');
    const d = anychart.data.set(data);
    for (let i = 0; i < showHeaders.length; i++) {
        const index = findColumnIndex(headers, showHeaders[i]);
        spline(c, d, index, showHeaders[i])
    }

    spliceDraw(c, 'container')
}

anychart.onDocumentReady(drawChart)
