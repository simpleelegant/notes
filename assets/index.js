var viewer = {
	template: '#viewer',
	props: ['id', 'title', 'html', 'diagramSVG'],
	methods: {
		onEdit: function() { this.$emit('edit') },
		onDelete: function() {
			var sure = prompt('enter DELETE to sure:','')
			if (sure === null) { return }
			if (sure !== 'DELETE') {
				alert('unexpected input')
				return
			}
			this.$http.post('/articles/delete?id='+this.id)
				.then(function(data) {
					this.$emit('deleted')
				}, function(data) { alert(data.bodyText) })
		},
		onMove: function() {
			var parent = prompt('enter an article id as new parent:','')
			if (parent === null) { return }
			if (parent === '') {
				alert('unexpected input')
				return
			}
			this.$http.post('/articles/update', {
				id: this.id,
				parent: parent,
				uParent: true
			}, {emulateJSON: true}).then(function(data) {
					this.$emit('moved')
				}, function(data) { alert(data.bodyText) })
		},
		onCreate: function() {
			if (!confirm('Create a sub-article?')) { return }
			this.$http.post('/articles/create', {
				parent: this.id,
				title: 'Article Title',
				content: 'Article Content',
			}, {emulateJSON: true}).then(function(data) {
					this.$emit('update:newArticleID', data.body.id)
				}, function(data) { alert(data.bodyText) })
		},
		onDraw: function() { this.$emit('draw') },
		onSearch: function() { this.$emit('search') },
		onExportRestore: function() { this.$emit('export-restore') }
	},
	updated: function() {
		// make highlight at source codes
		var pre = document.getElementsByTagName('pre')
		for (var i = pre.length - 1; i >= 0; i--) {
			var code = pre.item(i).getElementsByTagName('code')
			if (code.length) {
				hljs.highlightBlock(code.item(0))
			}
		}
	}
}

var editor = {
	template: '#editor',
	props: ['id', 'title', 'content', 'contentMD5'],
	data: function() {
		return {
			titleEditable: this.title,
			contentEditable: this.content
		}
	},
	methods: {
		onClose: function() {
			this.$http.post('/md5', {data: this.contentEditable}, {emulateJSON: true})
				.then(function(data) {
					if (data.body.md5 !== this.contentMD5
						&& !confirm('content was changed, discard?')) {
						return
					}
					this.$emit('close')
				}, function(data) { alert(data.bodyText) })
		},
		onSave: function() {
			this.save(false)
		},
		onSaveClose: function() {
			this.save(true)
		},
		save: function(close) {
			this.$http.post('/articles/update', {
				id: this.id,
				originalContentMD5: this.contentMD5,
				title: this.titleEditable,
				content: this.contentEditable,
				uTitle: true,
				uContent: true
			}, {emulateJSON: true}).then(function(data) {
				this.$emit('updated')
				if (close) {
					this.$emit('close')
				}
			}, function(data) { alert(data.bodyText) })
		}
	},
	mounted: function() {
		document.getElementsByTagName('textarea').item(0).focus()
	}
}

var articleView = {
	template: '#article-view',
	props: ['article'],
	components: {
		viewer: viewer,
		editor: editor
	},
	data: function() {
		return {
			newArticleID: '',
			edit: false,
			drawDiagram: false,
			parent: { id: '', title: '' },
			childrenOfParent: [],
			current: {
				id: '',
				title: '',
				content: '',
				html: '',
				diagram: '',
				diagramSVG: '',
				contentMD5: '',
				diagramMD5: ''
			},
			childrenOfCurrent: []
		}
	},
	methods: {
		show: function(id) {
			if (this.edit) {
				if (!confirm('Discard your editing?')) { return }
				this.edit = false
			}
			this.load(id)
		},
		onMoved: function() { this.load(this.current.id) },
		onUpdated: function() { this.load(this.current.id) },
		onDeleted: function() { this.load(this.parent.id) },
		load: function(articleID, edit) {
			this.$http.get('/articles/get?id='+articleID).then(function(data) {
				this.current = data.body.current
				this.parent = data.body.parent
				this.childrenOfParent = data.body.childrenOfParent ||
					[{id: this.current.id, title: this.current.title}]
				this.childrenOfCurrent = data.body.childrenOfCurrent

				if (edit) { this.edit = true }
			}, function(data) {
				this.error = data.url+': '+data.bodyText
			})
		},
		onCloseDiagramEditor: function() {
			this.$http.post('/md5', {data: this.current.diagram}, {emulateJSON: true})
				.then(function(data) {
					if (data.body.md5 !== this.current.diagramMD5
						&& !confirm('Diagram was changed, discard?')) {
						return
					}
					this.drawDiagram = false
					this.load(this.current.id)
				}, function(data) { alert(data.bodyText) })
		},
		onSaveDiagram: function() {
			this.$http.post('/articles/update', {
				id: this.current.id,
				originalDiagramMD5: this.current.diagramMD5,
				diagram: this.current.diagram,
				uDiagram: true
			}, {emulateJSON: true}).then(function(data) {
				this.load(this.current.id)
			}, function(data) { alert(data.bodyText) })
		},
		onSearch: function() { this.$emit('search') },
		onExportRestore: function() { this.$emit('export-restore') }
	},
	created: function() {
		this.load(this.article || '')
		this.$watch('newArticleID', function (newValue, oldValue) {
			this.load(this.newArticleID, true)
		})

		this.$watch('current.diagram', function (newValue, oldValue) {
			if (!newValue) {
				this.current.diagramSVG = ''
				return
			}
			this.$http.post('/diagram/render', {source: newValue}, {emulateJSON: true})
				.then(function(data) {
					this.current.diagramSVG = data.body.svg
				}, function(data) {
					this.current.diagramSVG = data.bodyText
				})
		})
	}
}

var search = {
	template: '#search',
	data: function() {
		return {
			tips: '',
			pattern: '',
			titleMatches: null,
			contentMatches: null
		}
	},
	methods: {
		onBack: function() { this.$emit('back') },
		onSubmit: function() {
			this.pattern = this.pattern.trim()
			if (this.pattern === '') { return }
			this.$http.post('/articles/search', {pattern: this.pattern}, {emulateJSON: true})
				.then(function(data) {
					this.titleMatches = data.body.titleMatches
					this.contentMatches = data.body.contentMatches
					this.tips = (!this.titleMatches && !this.contentMatches) ? 'No article matched' : ''
				}, function(data) { alert(data.bodyText) })
		},
		load: function(id) { this.$emit('update:article', id) }
	},
	mounted: function() {
		document.getElementsByTagName('input').item(0).focus()
	}
}

var exportRestore = {
	template: '#export-restore',
	methods: {
		onBack: function() { this.$emit('back') }
	}
}

new Vue({
	el: '#body',
	components: {
		'article-view': articleView,
		'search': search,
		'export-restore': exportRestore
	},
	data: function() {
		return {
			page: 'article',
			article: ''
		}
	},
	methods: {
		load: function(article) {
			this.article = article
			this.page = 'article'
		}
	}
})
