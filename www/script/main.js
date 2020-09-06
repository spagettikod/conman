import { ContainerCard } from './ContainerCard.js'

var app;

function init() {
    app = new Vue({
        el: '#app',
        data: {
            containers: [],
            filterValue: ''
        },
        components: {
            'container-card': ContainerCard
        },
        computed: {
            'filteredContainers': function () {
                let v = this.filterValue.toLowerCase();
                return this.containers.filter(function (c) {
                    if (v === '') {
                        return c;
                    }
                    if (c.name.toLowerCase().indexOf(v) > -1) {
                        return c;
                    }
                    if (c.image.toLowerCase().indexOf(v) > -1) {
                        return c;
                    }
                    if (c.state.toLowerCase().startsWith(v)) {
                        return c;
                    }
                });
            }
        },
        beforeMount: async function () {
            this.load();
        },
        mounted: function () {
            window.setInterval(this.load, 1000);
        },
        methods: {
            filterChanged: function (e) {
                if (e.keyCode === 27) {
                    e.target.value = '';
                }
            },
            stateClass: function (state) {
                switch (state) {
                    case 'running':
                        return 'badge-success';
                    case 'created':
                        return 'badge-warning';
                    case 'paused':
                        return 'badge-warning';
                    case 'exited':
                        return 'badge-danger';
                    case 'dead':
                        return 'badge-danger';
                    default:
                        return 'badge-light';
                }
            },
            load: async function () {
                let response = await fetch('/api/containers');
                this.containers = await response.json();
            },
            action: async function (link) {
                let response = await fetch(link.href, { method: link.type });
                if (response.ok) {
                    this.load();
                    return;
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