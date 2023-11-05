import os
import tempfile
from django.shortcuts import render
from django.http import HttpResponse, HttpResponseRedirect


def main(request):
    if not request.user.is_authenticated:
        return HttpResponseRedirect("/login/?next=/svsignal/trend/")

    context = {}
    if request.GET:
        context["Signals"] = request.GET.getlist("signalkey")
        context["Begin"] = request.GET["begin"]
        context["End"] = request.GET["end"]
        context["Title"] = request.GET["title"]
        context["Cols"] = request.GET["cols"]
        context["Height"] = request.GET["height"]
        context["GroupChart"] = request.GET["GroupChart"]

    return render(request, 'svsignal_dj_app/svsignal.html', context)
