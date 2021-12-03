var Signals = []
var Begin = 0
var End = 0

function draw_trend(series) {
    Highcharts.chart('trend', {
    /*
        chart: {
            zoomType: 'x'
        },
    */
        title: {
            text: 'График'
        },
    
        /*subtitle: {
            text: 'Using the Boost module'
        },*/
    /*
        tooltip: {
            valueDecimals: 2
        },
    */
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
            }
        },
        
        time: {
            //timezone: 'Asia/Novokuznetsk',
            timezoneOffset: -7*60,
        },

        legend: {
            layout: 'vertical',
            align: 'right',
            verticalAlign: 'middle'
        },
        
        plotOptions: {
            series: {
                label: {
                    connectorAllowed: false
                },
                pointStart: 1638319298
            }
        },

        series: series
    
    });
}

$( document ).ready(function() {
    let series = [];
    /*series: [{
        data: data,
        lineWidth: 0.5,
        name: 'Hourly data points'
    }]*/

    for(let i=0;i<Signals.length;i++) {
        let url = `/svs/api/signal/getdata?signalkey=${Signals[i]}&begin=${Begin}&end=${End}`;
        $.ajax({
            context: this,
            url: url,
            type: 'GET',
            success: function (result) {
                let data = [];
                if(result.values != undefined) {
                    for(let j=0;j<result.values.length;j++) {
                        // result.values[j][3] - offline
                        data.push([
                                result.values[j][1] * 1000,
                                result.values[j][2]
                            ]
                        )
                    }     
                }
                series.push({
                    data: data,
                    lineWidth: 0.5,
                    name: Signals[i]
                })
                if(i==Signals.length-1) {
                    draw_trend(series);
                }
            }.bind(this),
            error: function (jqXHR, exception) {
                if(i==Signals.length-1) {
                    draw_trend(series);
                }
                console.log(i, "ERROR", jqXHR, exception);
            }
        });
    }
    
   
});


/*

    $.ajax({
      context: this,
      url: '/aiohttp/MDTTruck/api/rb_category/',
      type: 'GET',
      success: function (result) {
        for(let idx=0; idx < result.length; idx++) {
          document.category[result[idx].id] = {name: result[idx].name, sub: {}};
        }
        $.ajax({
          context: this,
          url: '/aiohttp/MDTTruck/api/rb_sub_category/0/',
          type: 'GET',
          success: function (result) {
            for(let idx=0; idx < result.length; idx++) {
              document.category[result[idx].category_id].sub[result[idx].id] = result[idx].name;
            }
            // обновить в открытом окне у select'ов опции
            for(let idx=0; idx < document.timelineEdit.events.length; idx++) {
              initSelectCategory(document.timelineEdit.events[idx]);
            }
          }.bind(this),
          error: function (jqXHR, exception) {
            new Noty({
              timeout: 10000,
              queue: 'emergency',
              layout: 'bottomRight',
              type: 'error',
              theme: 'bootstrap-v4',
              text: 'Ошибка. ' + '(' + parse_error_ajax(jqXHR, exception) + ")"
            }).show();
          }
        });
      }.bind(this),
      error: function (jqXHR, exception) {
        new Noty({
          timeout: 10000,
          queue: 'emergency',
          layout: 'bottomRight',
          type: 'error',
          theme: 'bootstrap-v4',
          text: `Ошибка. (${parse_error_ajax(jqXHR, exception)})`
        }).show();
      }
    });

*/