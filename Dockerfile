FROM quay.io/mocaccino/extra
ENV LUET_YES=true
RUN luet install container/img vcs/git 
