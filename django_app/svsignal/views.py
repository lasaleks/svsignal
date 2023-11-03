import os
import tempfile
from django.shortcuts import render
from django.http import HttpResponse, HttpResponseRedirect


def index(request):
    if not request.user.is_authenticated:
        return HttpResponseRedirect("/login/?next=/svsignal/")
    return render(request, 'svsignal_django/svsignal.html', context={})
