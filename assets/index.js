// wrap all global variables & functions
window.A = {
    parseQueryString: function() {
        var qs = location.search.length ? location.search.substr(1).split('&') : [],
            args = {};

        qs.forEach(function(q) {
            if (q) {
                var kv = q.split('=');
                if (kv.length === 2) {
                    args[decodeURIComponent(kv[0])] = decodeURIComponent(kv[1]);
                }
            }
        });

        return args;
    },

    request: function (method, url, data, success, fail) {
        $.ajax({
            method: method, // 'GET', 'POST', 'PUT', etc.
            url: url,
            data: data,
            success: success,
            error: function (jqXHR) {
                if (jqXHR.status === 401) {
                    alert('authentication come soon');
                } else if (fail) {
                    fail(jqXHR.status,jqXHR.statusText,jqXHR.responseJSON,jqXHR.responseText);
                } else {
                    alert((jqXHR.responseJSON && jqXHR.responseJSON.message) ? 
                        jqXHR.responseJSON.message :
                            (jqXHR.responseText || jqXHR.status+' '+jqXHR.statusText));
                }
			}
        });
    },

    // sup = with super article title
    // sub = with sub-articles (just id & title)
    // subling = with subling articles (just id & title)
    // html = render article content in HTML 
    getArticle: function(id, sup, sub, subling, html, success, fail) {
        this.request('GET',
            '/articles/get',
            { id: id, sup: sup, sub: sub, subling: subling, html: html },
            success,
            fail);
    },

    searchArticles: function(title, success) {
        this.request('GET', '/articles/search', { title: title }, success);
    },

    renderDiagram: function(source, success, fail) {
        this.request('POST', '/diagram/render', {source: source}, success, fail);
    },

	getMD5: function(data, success) {
		this.request('POST', '/md5', {data: data}, function(data) { success(data.md5); });
	},

    loadArticle: function(id, dontPushURL) {
        this.getArticle(id, true, true, true, true, function(a) {
            A.article = a;

            // update location bar of browser
            if (!dontPushURL && history.pushState) {
                history.pushState(null, '', location.pathname + '?id=' + id);
            }

            // update title bar of browser
            document.title = a.title + ' - notes';

            $('#title').text(a.title);
            $('#content > div').html(a.content_in_html);
            $('pre code').each(function(i, block) { hljs.highlightBlock(block); /* highlight on code block */ });

            // render super article link
            if (a.parent_id) {
                $('#super-article').text(a.super_article_title).attr('data-id', a.parent_id);
            } else {
                $('#super-article').text('').attr('data-id', '');
            }

            // render sub-article list
            var $subArticles = $('#sub-articles > div').empty();
            if (a.sub_articles && a.sub_articles.length) {
                a.sub_articles.forEach(function(a) {
                    $('<a onclick="A.clickArticleLink(this);"></a>').text(a.title).attr('data-id', a.id).appendTo($subArticles);
                });
            } else {
                $subArticles.append('<p style="text-align: center; color: grey;">no sub-articles.</p>');
            }

            // render sibling article list
            var $sibling = $('#sibling-articles > div').empty();
            if (a.sibling_articles && a.sibling_articles.length) {
                a.sibling_articles.forEach(function(b) {
                    if (b.id === a.id) { return; }
                    $('<a onclick="A.clickArticleLink(this);"></a>').text(b.title).attr('data-id', b.id).appendTo($sibling);
                });
            } else {
                $sibling.append('<p style="text-align: center; color: grey;">no sibling articles.</p>');
            }

            // make TOC
            A.makeTOC($('#content > div'), $('#topics > div'));
        });
    },

    loadArticleByURL: function() {
        this.loadArticle(this.parseQueryString().id || '', true);
    },

    clickArticleLink: function(a) {
        A.loadArticle($(a).attr('data-id'));
    },

    makeTOC: function($articleContentElement, $tocElement) {
        $tocElement.empty();

        $articleContentElement.find('h1,h2,h3,h4,h5,h6').each(function(i) {
            $(this).attr('id', 'h-' + i); // add "id" attribute
            $('<a style="margin-left:' + (this.tagName.charAt(1) - 1) + 'em;" href="#h-' + i + '"></a>').text(this.innerText).appendTo($tocElement);
        });
    }
};

/*
// view model
var vm = {
    addDoc: function() {
        if (!confirm('Add Sub-article?')) {
            return;
        }

        A.request('POST', '/articles/create', {
            parentID: this.id(),
            title: 'New Article',
            content: ''
        }, function(data) {
            vm.superDoc({ id: vm.id(), title: vm.title() });

            vm.id(data.id)
                .parentID(data.parentID)
                .title(data.title)
                .content(data.content)
				.contentMD5(data.contentMD5)
				.diagramMD5(data.diagramMD5);

            vm.subDocs([]).getSiblingDocs();

            // refresh location
            if (history.pushState) {
                history.pushState(null, '', location.pathname + '?id=' + data.id);
            }

            vm.viewMode(false).diagramMode(false).editContentMode(true);
        });
    },

    deleteDoc: function() {
        if (prompt('enter DELETE to sure:', '') !== 'DELETE') {
            return;
        }

        A.request('POST', '/articles/delete', { id: this.id() }, function(data) {
            vm.loadDoc(vm.parentID(), true);
        });
    },

    setParentID: function() {
        var parentID = prompt('Specify new parent article by id:', '');
        if (!parentID) { return; }

        A.request('POST', '/articles/update', { id: this.id(), parentID: parentID }, function(data) {
            vm.parentID(parentID);
            vm.getSuperDoc().getSiblingDocs();
        });
    },

    saveDiagram: function(success) {
        var dia = vm.diagram().trim();

        function save() {
            A.request('POST', '/articles/update', {
				id: vm.id(),
				diagram: dia,
				beforeDiagramMD5: vm.diagramMD5(),
			}, function(data) {
				vm.diagramMD5(data.diagramMD5);
                if (success) { success(); }
            });
        }

        if (dia === '') { save(); return; }

        A.parseDiagram(dia, function(data) {
            save();
        }, function(data) {
            if (confirm('Diagram syntax is invalid, save anyway?')) { save(); }
        });
    },

	cancelEditContent: function() {
		A.getMD5(vm.content(), function(md5) {
			if (md5 === vm.contentMD5() || confirm('Article content was changed. Cancel?')) {
				vm.editContentMode(false);
				vm.loadDoc(vm.id(), false);
			}
		});
	},

	cancelEditDiagram: function() {
		A.getMD5(vm.diagram(), function(md5) {
			if (md5 === vm.diagramMD5() || confirm('Diagram was changed. Cancel?')) {
				vm.editDiagramMode(false);
				vm.viewMode(true);
				vm.loadDoc(vm.id(), false);
			}
		});
	},

    delayClearSearching: function() {
        setTimeout(function() {
            vm.searchPattern('').searchResult([]);
        }, 400);
    },

    showTips: function(tips) {
        var $tips = $('.bottom-tips');

        if (!$tips.length) {
            $tips = $('<div class="bottom-tips" style="position: fixed; bottom: 0; z-index: 999; background-color: rgba(0, 0, 0, 0.71); color: #fff; padding: 4px 10px; border-radius: 0 8px 0 0; font-size: 120%; left: 0; display: none;"></div>');
            $tips.appendTo($('body'));
        }

        $tips.text(tips).fadeIn();

        return { then: function(callback) { callback($tips); } };
    }
};

vm.viewMode.subscribe(function(newValue) {
    if (newValue) {
        //setTimeout(function() {
            $('#bac').find('a').attr('target', '_blank');
            $('pre code').each(function(i, block) {
                hljs.highlightBlock(block);
            });
            A.generateCatalog('#bac');
        //}, 400);
    }
});

vm.diagram.subscribe(function() {
    if (vm.updateDiagramTimeout) { return; }

    vm.updateDiagramTimeout = setTimeout(function() {
        clearTimeout(vm.updateDiagramTimeout);
        vm.updateDiagramTimeout = null;

        var dia = vm.diagram().trim();
        if (dia === '') {
            $('.diagram').empty().append('no diagram.');
            return;
        }

        A.parseDiagram(dia, function(data) {
            $('.diagram').empty().append(data.svg);
        }, function(data) {
			$('.diagram').empty().append('diagram error: '+
				(data.responseJSON && data.responseJSON.message) ? data.responseJSON.message : data.statusText);
        });
    }, 1000);
});
*/

