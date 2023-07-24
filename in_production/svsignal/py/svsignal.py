import json

def set_attr_signal(signalkey, name, typesave, period, delta, tags):
    exchange = 'svsignal'
    routing_key = f"svs.set.{signalkey}"
    json_data = json.dumps({
        'typesave': typesave,
        'period': period,
        'delta': delta,
        'name': name,
        'tags': tags
    })
    return exchange, routing_key, json_data


def save_value(signalkey, value, utime, offline):
    exchange = 'svsignal'
    routing_key = f"svs.save.{signalkey}"
    json_data = json.dumps({
        'value': value,
        'utime': utime,
        'offline': offline
    })
    return exchange, routing_key, json_data
