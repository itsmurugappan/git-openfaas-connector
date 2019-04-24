FROM scratch
ADD release/git-faas /
CMD ["/git-faas"]
