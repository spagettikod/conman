var app;

function init() {
    app = new Vue({
        el: '#app',
        data: {
            services: [],
            filterValue: ''
        },
        computed: {
            'filteredServices': function () {
                let v = this.filterValue.toLowerCase();
                return this.services.filter(function (svc) {
                    if (v === '') {
                        return svc;
                    }
                    if (svc.name.toLowerCase().indexOf(v) > -1) {
                        return svc;
                    }
                    if (svc.image.toLowerCase().indexOf(v) > -1) {
                        return svc;
                    }
                });
            }
        },
        beforeMount: async function () {
            this.loadServices();
        },
        mounted: function () {
            window.setInterval(this.loadServices, 3000);
        },
        methods: {
            filterChanged: function (e) {
                if (e.keyCode === 27) {
                    e.target.value = '';
                }
            },
            loadServices: async function () {
                let response = await fetch('/api/services');
                if (response.ok) {
                    this.services = await response.json();
                }
            },
            downloadLog: async function (link) {
                response = await fetch(link.href, { method: link.type });
                if (response.ok) {
                    this.log = await response.text()
                    return;
                }
            }
        }
    });
}

window.addEventListener('load', init);