function draw_trend(series, max_y, min_y, min_x, max_x) {
    let options = {
        chart: {
            height: (9 / 16 * 80) + '%', // 16:9 ratio
            type: 'line',
            zoomType: 'xy'
        },
    
        title: {
            text: null
        },

        yAxis: {
            title: {
                text: 'Значение'
            },
            scrollbar: {
                enabled: true
            }
        },
        xAxis: {
            useUTC: false,
            type: 'datetime',
            title: {
                text: 'Время'
            },
            scrollbar: {
                enabled: true
            },
            startOnTick: false,
            endOnTick: false,
            startOnTick: false,
            min: min_x,
            max: max_x,

        },
        time: {
            //timezone: 'Asia/Novokuznetsk',
            timezoneOffset: -7*60,
        },
        legend: {
            /*layout: 'vertical',
            align: 'right',
            verticalAlign: 'middle'*/
        },
        plotOptions: {
            series: {
                lineWidth: 1.3,
                //pointPlacement: 'on'
            }
        },
        series: series,

    };
    if(max_y!==undefined) {
        options.yAxis.max = max_y;
    }
    if(min_y!==undefined) {
        options.yAxis.min = min_y;
    }
    Highcharts.chart('trend', options);
}

/*
function draw_trend(series, max_y, min_y, min_x, max_x) {
    console.log("draw_trend", max_y, min_y, min_x, max_x)
    let options = {
        chart: {
            //type: 'line',
            height: (9 / 16 * 80) + '%', // 16:9 ratio
        },
        yAxis: {
            title: {
                text: 'Значение'
            }
        },
        xAxis: {
            type: 'datetime',
            useUTC: false,
            title: {
                text: 'Время'
            },
            min: min_x,
            max: max_x,
        },
        rangeSelector: {
            enabled:false,
            //inputEnabled: false
        },
        legend: {
            layout: 'vertical',
            align: 'right',
            verticalAlign: 'middle'
        },
        plotOptions: {
            series: {
                showInNavigator: true,
            }
        },
        legend: {
            enabled: true,
            layout: 'vertical',
            align: 'right',
            verticalAlign: 'top',
            y: 30,
            navigation: {
                activeColor: '#3E576F',
                animation: true,
                arrowSize: 12,
                inactiveColor: '#CCC',
                style: {
                    fontWeight: 'bold',
                    color: '#333',
                    fontSize: '12px'
                }
            }
        },
        stockTools:{
            gui:{
                buttons:[ 
                    //'indicators',
                    //'separator', 
                    'simpleShapes', 
                    //'lines', 
                    //'crookedLines', 
                    //'measure', 
                    //'advanced', 
                    //'toggleAnnotations', 
                    //'separator', 
                    //'verticalLabels', 
                    //'flags', 
                    //'separator', 
                    'zoomChange', 
                    'fullScreen', 
                    //'typeChange', 
                    //'separator', 
                    //'currentPriceIndicator', 
                    //'saveChart' 
                ],
                //className: "highcharts-bindings-wrapper",
                //definitions:{},
                //enabled:true,
                //iconsURL:undefined,
                //toolbarClassName:"stocktools-toolbar",
            }
        },
        series: series,
    };
    if(max_y!==undefined) {
        options.yAxis.max = max_y;
    }
    if(min_y!==undefined) {
        options.yAxis.min = min_y;
    }
    console.log(options);
    Highcharts.stockChart('trend', options);
}*/


function strDate2Date(str_date) {
    let _date = str_date.split(' ')[0];
    let _time = str_date.split(' ')[1];
    let day = parseInt(_date.split('.')[0]), mount = parseInt(_date.split('.')[1])-1, year = parseInt(_date.split('.')[2]);
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

$( document ).ready(function() {
    let dtbegin = new Date(Begin*1000);
    let dtend = new Date(End*1000);

    let pbegin = $('#begin').datetimepicker({
        step:10,
        mask:true,
        format:'d.m.Y H:i:s',
        onChangeDateTime:function(dp) {
            dtbegin = dp;
        }
    });
    pbegin.val(DateToStrDate(dtbegin));
    
    let pend = $('#end').datetimepicker({
        step:10,
        mask:true,
        format:'d.m.Y H:i:s',
        onChangeDateTime:function(dp) {
            dtend = dp;
        }
    });
    pend.val(DateToStrDate(dtend));

    queryDataTrend(parseInt(dtbegin.getTime() / 1000), parseInt(dtend.getTime() / 1000), Signals);
       
    document.getElementById("btnQuery").addEventListener("click", e => {
        queryDataTrend(parseInt(dtbegin.getTime() / 1000), parseInt(dtend.getTime() / 1000), Signals);
    });

    function queryDataTrend(begin, end, Signals) {
        let series = [];
        let cnt = 0;
        let max_y = undefined, min_y = undefined;
        let query_idx = 0
        for(let i=0;i<Signals.length;i++) {
            let url = `/svs/api/signal/getdata?signalkey=${Signals[i]}&begin=${begin}&end=${end}`;
            $.ajax({
                context: this,
                url: url,
                type: 'GET',
                success: function (result) {
                    console.log('result:', result);
                    for(let i=0;i<result.tags.length;i++) {
                        switch(result.tags[i].tag) {
                            case 'max_y': {
                                let value = parseFloat(result.tags[i].value); 
                                if(max_y === undefined) {
                                    max_y = value;
                                } else {
                                    max_y = value > max_y ? value : max_y;
                                }
                            }
                            break;
                            case 'min_y': {
                                let value = parseFloat(result.tags[i].value);
                                if(min_y === undefined) {
                                    min_y = value;
                                } else {
                                    min_y = value < min_y ? value : min_y;
                                }
                            }
                            break;
                        }
                    }
                    let data = [];
                    if(result.values != undefined) {
                        switch(result.typesave) {
                        case 2:
                            for(let j=0;j<result.values.length;j++) {
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
                            for(let j=0;j<result.values.length;j++) {
                                let utime = result.values[j][1] * 1000;
                                let value = result.values[j][2];
                                if(prev_value != null && value != prev_value && utime != prev_utime) {
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
                        name: Signals[i],
                        tooltip: {
                            valueDecimals: 2
                        },
                        showInLegend: true,
                    })
                    if(query_idx==Signals.length-1) {
                        draw_trend(series, max_y, min_y, begin*1000, end*1000);
                    }
                    query_idx++;
                }.bind(this),
                error: function (jqXHR, exception) {
                    if(query_idx==Signals.length-1) {
                        draw_trend(series, max_y, min_y, begin*1000, end*1000);
                    }
                    query_idx++;
                    console.log(i, "ERROR", jqXHR, exception);
                }
            });
        }
    }
});
