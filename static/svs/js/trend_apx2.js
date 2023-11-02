var ru = null;
var signal_charts = [];

function create_chart(signal_key, y_fixed, series_name, series_data, series_color, site, max_y, min_y, min_x, max_x, ru) {
    // console.log("create_chart:", signal_key, series_name, series_color, max_y, min_y, min_x, max_x, ru, "series_data.length:", series_data.length);
    f_y_axis_labels = {
        formatter: function(val) {
            return val;
        }
    }

    if(y_fixed != null) {
        f_y_axis_labels = {
            formatter: function(val) {
                return val.toFixed(y_fixed);
            }
        }
    }

    var options = {
        series: [{
            data: series_data,
            name: series_name,
        }],
        chart: {
            animations: {
                enabled: false,
            },
            redrawOnWindowResize: true,
            redrawOnParentResize: false,
            locales: [ru],
            defaultLocale: 'ru',
            id: `${signal_key}_chart`,
            type: 'line',
            height: "100%",
            dropShadow: {
                enabled: true,
                color: '#000',
                top: 18,
                left: 7,
                blur: 10,
                opacity: 0.2
            },
            zoom: {
                type: 'xy',
                enabled: true,
                //autoScaleYaxis: true,
            },
            toolbar: {
                autoSelected: 'zoom',
            },
        },
        title: {
            text: series_name,
            offsetX: 35,
            floating: true,
            maring:10,
            style: {
              fontSize:  '14px',
              fontWeight:  'normal',
            },
        },        
        subtitle: {
            text: site,
            align: 'left',
            offsetX: 35,
            offsetY: 25,
            floating: false,
            style: {
                fontSize:  '14px',
                fontWeight:  'bold',
            },
        },
        stroke: {
            //curve: 'stepline',
            width: 3
        },
        fill: {
            opacity: 1,
        },
        legend: {
            show: true
        },
        xaxis: {
            type: "datetime",
            tickAmount: 10,
            labels: {
                datetimeFormatter: {
                    fulldate: "DD/MM",
                    day: "DD/MM",
                    hour: "HH:mm",
                    minute: "HH:mm:ss",
                    month: "YY/MM",
                    second: "HH:mm:ss",
                    year: "yyyy",
                },
                datetimeUTC: false,
                formatter: function (value, timestamp, opts) {
                    let diff = opts.w.globals.maxX - opts.w.globals.minX;
                    let dtf = opts.w.config.xaxis.labels.datetimeFormatter;
                    let format = dtf.second
                    if (diff > 31536000000 * 5) {
                        format = dtf.year
                    } else if (diff > 2592000000 * 5) {
                        `chart_${idx}`,
                        format = dtf.day
                    } else if (diff > 3600000 * 5) {
                        format = dtf.hour
                    } else if (diff > 60 * 5) {
                        format = dtf.minute
                    }
                    return moment.tz(timestamp, TimeZone).format(format)
                },
            },
            title: {
                text: "Время"
            },
            min: min_x,
            max: max_x,
        },
        yaxis: {
            title: {
                text: "Значение"
            },
            min: min_y,
            max: max_y,
            // decimalsInFloat: 1,
            labels: f_y_axis_labels,
        },
        tooltip: {
            x: {
                show: true,
                //format: 'DD/MM/yy hh:mm:ss',
                formatter: function (timestamp, w) {
                    return moment.tz(timestamp, TimeZone).format('DD/MM/yy HH:mm:ss')
                }
            },
        },
        grid: {
            row: {
                colors: ['#e5e5e5', 'transparent'],
                opacity: 0.5
            },
            column: {
                colors: ['#f8f8f8', 'transparent'],
            },
            xaxis: {
                lines: {
                    show: true
                }
            }
        },
    };
    
    if (UseGroupChart == 1 && series_data.length > 0) {
        options.chart.group = "general_timing"
    }
    if (series_color) {
        options.colors = [series_color,]
    }
    //console.log(options);
    let chart = new ApexCharts(document.getElementById(`${signal_key}`), options);
    chart.render();
    return chart
}

function update_chart(chart, y_fixed, signal_key, series_data, max_y, min_y, min_x, max_x) {
    // console.log("update_chart:", chart, "signal_key:", signal_key, "max_y:", max_y, "min_y:", min_y, "min_x:", min_x, "max_x:", max_x, "series_data.length:", series_data.length);
    if (series_data.length > 0) {
        //options.chart.group = "general_timing"
        chart.updateSeries([{ data: series_data }])
        chart.updateOptions(
            {
                xaxis: {
                    type: "datetime",
                    tickAmount: 10,
                    labels: {
                        datetimeFormatter: {
                            fulldate: "DD/MM",
                            day: "DD/MM",
                            hour: "HH:mm",
                            minute: "HH:mm:ss",
                            month: "YY/MM",
                            second: "HH:mm:ss",
                            year: "yyyy",
                        },
                        datetimeUTC: false,
                        formatter: function (value, timestamp, opts) {
                            let diff = opts.w.globals.maxX - opts.w.globals.minX;
                            let dtf = opts.w.config.xaxis.labels.datetimeFormatter;
                            let format = dtf.second
                            if (diff > 31536000000 * 5) {
                                format = dtf.year
                            } else if (diff > 2592000000 * 5) {
                                format = dtf.day
                            } else if (diff > 3600000 * 5) {
                                format = dtf.hour
                            } else if (diff > 60 * 5) {
                                format = dtf.minute
                            }
                            return moment.tz(timestamp, TimeZone).format(format)
                        },
                    },
                    title: {
                        text: "Время"
                    },
                    min: min_x,
                    max: max_x,
                },
                /*yaxis: {
                    title: {
                        text: "Значение"
                    },
                    min: min_y,
                    max: max_y,
                    // decimalsInFloat: 1,
                    labels: {
                        formatter: function(val) {
                            return val;
                        }
                    },
                },*/
                chart: {
                    group: "general_timing",
                },
            },
            true,
            true,
            true
        );
    }
}

function draw_trend(series, min_x, max_x, ru) {
    for (let i = 0; i < series.length; i++) {
        let sigchart = signal_charts[series[i].signal_key];
        if (sigchart === undefined) {
            sigchart = create_chart(series[i].signal_key, series[i].y_fixed, series[i].name, series[i].data, series[i].series_color, series[i].site, series[i].max_y, series[i].min_y, min_x, max_x, ru);
            signal_charts[series[i].signal_key] = sigchart;
        } else {
            update_chart(sigchart, series[i].y_fixed, series[i].signal_key, series[i].data, series[i].max_y, series[i].min_y, min_x, max_x);
        }
    }
}

function strDate2Date(str_date) {
    let _date = str_date.split(' ')[0];
    let _time = str_date.split(' ')[1];
    let day = parseInt(_date.split('.')[0]), mount = parseInt(_date.split('.')[1]) - 1, year = parseInt(_date.split('.')[2]);
    let hour = parseInt(_time.split(':')[0]), minute = parseInt(_time.split(':')[1]), second = parseInt(_time.split(':')[2]);
    return new Date(year, mount, day, hour, minute, second);
}

function DateToStrDate(date) {
    let dd = date.getDate();
    let MM = date.getMonth();
    let YYYY = date.getFullYear();
    let hh = date.getHours();
    let mm = date.getMinutes();
    let ss = date.getSeconds();
    return (dd > 9 ? '' : '0') + dd + '.' + (MM > 8 ? '' : '0') + (MM + 1) + '.' + YYYY + " " + (hh < 10 ? '0' : '') + hh + ':' + (mm < 10 ? '0' : '') + mm + ":" + (ss < 10 ? '0' : '') + ss;
}

$(document).ready(function () {
    let dtbegin = new Date(Begin * 1000);
    let dtend = new Date(End * 1000);

    let pbegin = $('#begin').datetimepicker({
        step: 10,
        mask: true,
        format: 'd.m.Y H:i:s',
        onChangeDateTime: function (dp) {
            dtbegin = dp;
        }
    });
    pbegin.val(DateToStrDate(dtbegin));

    let pend = $('#end').datetimepicker({
        step: 10,
        mask: true,
        format: 'd.m.Y H:i:s',
        onChangeDateTime: function (dp) {
            dtend = dp;
        }
    });
    pend.val(DateToStrDate(dtend));

    function create_charts_node() {
        let col = 0;
        let container = document.getElementById("container");
        Array(parseInt(Signals.length / Cols)+1).fill(Array(parseInt(Cols)))
        let rows = [];
        let nofrows = parseInt(Signals.length / Cols) + (Signals.length % Cols > 0 ? 1 : 0);
        //console.log("nofrows:", nofrows);
        //console.log("Signals.length:", Signals.length);
        for (let i = 0; i < nofrows; i++) {
            rows[i] = [];
            for (let j = 0; j < Cols; j++) {
                rows[i][j] = []
            }
        }

        for (let i = 0; i < nofrows; i++) {
            let row = document.createElement('div');
            row.classList.add("row");
            row.setAttribute("style", `height:${Height}vh`)
            for (let j = 0; j < Cols; j++) {
                let col = document.createElement('div');
                col.classList.add("col");
                rows[i][j] = col;
                row.appendChild(col);
            }
            container.appendChild(row);
        }
        //console.log(rows);

        for (let i = 1; i < Signals.length+1; i++) {
            let row = parseInt(i / Cols) + (i % Cols > 0 ? 1 : 0) - 1;
            let graph = document.createElement("div")
            graph.id = Signals[i-1];
            rows[row][col].appendChild(graph);
            if (col >= Cols - 1) col = 0
            else col++;
        }
    };
    create_charts_node();

    queryDataTrend(parseInt(dtbegin.getTime() / 1000), parseInt(dtend.getTime() / 1000), Signals);

    document.getElementById("btnQuery").addEventListener("click", e => {
        queryDataTrend(parseInt(dtbegin.getTime() / 1000), parseInt(dtend.getTime() / 1000), Signals);
    });

    function queryDataTrend(begin, end, Signals) {
        let series = [];
        let cnt = 0;
        let query_idx = 0
        for (let i = 0; i < Signals.length; i++) {
            let url = `/svs/api/signal/getdata?signalkey=${Signals[i]}&begin=${begin}&end=${end}`;
            $.ajax({
                context: this,
                url: url,
                type: 'GET',
                success: function (result) {
                    let max_y = undefined, min_y = undefined;
                    let series_color = undefined;
                    let site = "";
                    let y_fixed = null;
                    for (let i = 0; i < result.tags.length; i++) {
                        switch (result.tags[i].tag) {
                            case 'max_y': {
                                let value = parseFloat(result.tags[i].value);
                                if (max_y === undefined) {
                                    max_y = value;
                                } else {
                                    max_y = value > max_y ? value : max_y;
                                }
                            }
                            break;
                            case 'min_y': {
                                let value = parseFloat(result.tags[i].value);
                                if (min_y === undefined) {
                                    min_y = value;
                                } else {
                                    min_y = value < min_y ? value : min_y;
                                }
                            }
                            break
                            case 'series_color':
                                series_color = result.tags[i].value;
                            break;
                            case 'site':
                                site = result.tags[i].value;
                            break
                            case 'y_fixed':
                                y_fixed = result.tags[i].value;
                            break
                        }
                    }
                    let data = [];
                    if (result.values != undefined) {
                        switch (result.typesave) {
                            case 2:
                                for (let j = 0; j < result.values.length; j++) {
                                    // result.values[j][3] - offline
                                    data.push([
                                        result.values[j][1] * 1000,
                                        result.values[j][2]
                                    ]
                                    )
                                }
                                break;
                            case 1:
                                let prev_value = null;
                                let prev_utime = null;
                                for (let j = 0; j < result.values.length; j++) {
                                    let utime = result.values[j][1] * 1000;
                                    let value = result.values[j][2];
                                    if (prev_value != null && value != prev_value && utime != prev_utime) {
                                        data.push([utime, prev_value])
                                    }
                                    data.push([utime, value])
                                    prev_utime = utime;
                                    prev_value = value
                                }
                                break;
                        }
                    }
                    series.push({
                        data: data,
                        //lineWidth: 0.5,
                        signal_key: Signals[i],
                        name: result.signalname,
                        y_fixed: y_fixed,
                        max_y: max_y,
                        min_y: min_y,
                        series_color: series_color,
                        site: site,
                    })
                    if (query_idx == Signals.length - 1) {
                        $.getJSON('/static/apexcharts/locales/ru.json', function (ru) {
                            draw_trend(series, begin * 1000, end * 1000, ru);
                        });
                    }
                    query_idx++;
                }.bind(this),
                error: function (jqXHR, exception) {
                    if (query_idx == Signals.length - 1) {
                        $.getJSON('/static/apexcharts/locales/ru.json', function (ru) {
                            draw_trend(series, begin * 1000, end * 1000, ru);
                        });
                    }
                    query_idx++;
                    console.log(i, "ERROR", jqXHR, exception);
                }
            });
        }
    }
});
