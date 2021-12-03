
function draw_trend(series, max_y, min_y, min_x, max_x) {
    let options = {
        chart: {
            height: (9 / 16 * 100) + '%', // 16:9 ratio
            type: 'line',
            zoomType: 'xy'
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
            zoomEnabled: true,
            min: min_x,
            max: max_x,
            startOnTick: false
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
                lineWidth: 1,
                //pointPlacement: 'on'
            }
        },
        series: series
    };
    if(max_y!==undefined) {
        options.yAxis.max = max_y;
    }
    if(min_y!==undefined) {
        options.yAxis.min = min_y;
    }
    console.log(options);
    Highcharts.chart('trend', options);
}


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
    let now = new Date();
    let dtbegin = new Date();
    dtbegin.setHours(0);
    dtbegin.setMinutes(0);
    dtbegin.setSeconds(0);
    let dtend = new Date(dtbegin);
    dtend.setDate(dtend.getDate() + 1);

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
    
    $.ajax({
        context: this,
        url: '/svs/api/getlistsignal',
        type: 'GET',
        success: function (result) {
            init_treeview(result);
        }.bind(this),
        error: function (jqXHR, exception) {
        }
    });
    
    function init_treeview(data) {
        let tree = [];
        let t = 0;
        let groups = {};
        for (const [key, group] of Object.entries(data)) {
            console.log(group)
            groups[key] = {'group': group, 'site': {}}
            for (const [s_key, signal] of Object.entries(group.signals)) {
                site = 'other';
                for(let i=0;i<signal.tags.length;i++) {
                    switch(signal.tags[i].tag) {
                        case 'site':
                            site = signal.tags[i].value;
                        break;
                    }
                }
                signal_node = {'key': s_key, 'signal': signal}
                if (groups[key].site[site] === undefined) {
                    groups[key].site[site] = [signal_node,]
                } else {
                    groups[key].site[site].push(signal_node)
                }
            }
        }
        
        for (const [groupkey, group] of Object.entries(groups)) {
            nodes = [];
            other_signal = group.site['other'];
            nodes_other = [];
            for(let i = 0;i<other_signal.length;i++) {
                let name = other_signal[i].signal.name == '' ? other_signal[i].key : other_signal[i].signal.name;
                nodes_other.push({
                    id: other_signal[i].key, 
                    "text": `<i class="far fa-square checkbox"></i> ${name}`,
                    class:"signal"
                    })
            }
            nodes.push({id: `${groupkey}_other`, "text": '', 'nodes': nodes_other, icon: "fas fa-wave-square"})

            delete group.site['other'];
            for (const [namesite, signals] of Object.entries(group.site)) {
                signal_nodes = []
                for(let i = 0;i<signals.length;i++) {
                    let name = signals[i].signal.name == '' ? signals[i].key : signals[i].signal.name;
                    signal_nodes.push({
                        id: signals[i].key, 
                        "text": `<i class="far fa-square checkbox"></i> ${name}`, 
                        class:"signal"
                    })
                }
                nodes.push({id: `${groupkey}_other`, "text": namesite, 'nodes': signal_nodes, icon: "fas fa-wave-square"})
            }
            tree.push({id: groupkey, "text": ` ${groupkey} ${group.group.name}`, 'nodes': nodes, icon: "far fa-folder"})
        }

        $('#tree').bstreeview({ data: tree });
        document.querySelectorAll('.signal .checkbox').forEach(i => i.addEventListener(
            "click",
            e => {
                e.srcElement.classList.toggle('fa-square');
                e.srcElement.classList.toggle('fa-check-square');
            }
        )
        );
    }
    document.getElementById("btnQuery").addEventListener("click", e => {
        let signals = []
        document.querySelectorAll(".signal .fa-check-square").forEach(i => {
            signals.push(i.parentElement.id);
        });
        queryDataTrend(parseInt(dtbegin.getTime() / 1000), parseInt(dtend.getTime() / 1000), signals);
    });

    function queryDataTrend(begin, end, Signals) {
        let series = [];
        let cnt = 0;
        let max_y = undefined, min_y = undefined;
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
                        name: Signals[i]
                    })
                    if(i==Signals.length-1) {
                        draw_trend(series, max_y, min_y, begin*1000, end*1000);
                    }
                }.bind(this),
                error: function (jqXHR, exception) {
                    if(i==Signals.length-1) {
                        draw_trend(series, max_y, min_y, begin*1000, end*1000);
                    }
                    console.log(i, "ERROR", jqXHR, exception);
                }
            });
        }
    }
});
