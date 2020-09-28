import { ServiceCard } from './ServiceCard.js'
import { ContainerCard } from './ContainerCard.js'

var app;

function init() {
    app = new Vue({
        el: '#app',
        data: {
            services: [],
            containers: [],
            filterValue: '',
            settings: {
                autoUpdate: false,
                swarmMode: false
            },
            intervalID: null
        },
        components: {
            'service-card': ServiceCard,
            'container-card': ContainerCard
        },
        watch: {
            'settings.autoUpdate': function (newVal, oldVal) {
                this.saveSettings();
                if (newVal) {
                    if (!this.intervalID) {
                        this.intervalID = window.setInterval(this.loadData, 1000);
                    }
                } else {
                    if (this.intervalID) {
                        window.clearInterval(this.intervalID);
                        this.intervalID = null;
                    }
                }
            },
            'settings.swarmMode': function (newVal, oldVal) {
                this.saveSettings();
                this.loadData();
            }
        },
        computed: {
            'filteredContainers': function () {
                let v = this.filterValue.toLowerCase();
                return this.containers.filter(function (cont) {
                    if (v === '') {
                        return cont;
                    }
                    if (cont.name.toLowerCase().indexOf(v) > -1) {
                        return cont;
                    }
                    if (cont.image.toLowerCase().indexOf(v) > -1) {
                        return cont;
                    }
                    if (cont.state.toLowerCase().indexOf(v) > -1) {
                        return cont;
                    }
                });
            },
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
            this.loadData();
        },
        mounted: function () {
            this.loadSettings();
        },
        methods: {
            saveSettings: function () {
                localStorage.setItem('conman_setting_auto_update', this.settings.autoUpdate);
                localStorage.setItem('conman_setting_swarm_mode', this.settings.swarmMode);
            },
            loadSettings: function () {
                let autoUpdate = localStorage.getItem('conman_setting_auto_update') === 'true';
                let swarmMode = localStorage.getItem('conman_setting_swarm_mode') === 'true';
                this.settings.autoUpdate = autoUpdate;
                this.settings.swarmMode = swarmMode;
            },
            filterChanged: function (e) {
                if (e.keyCode === 27) {
                    e.target.value = '';
                }
            },
            loadData: async function () {
                if (this.settings.swarmMode) {
                    let response = await fetch('api/services');
                    if (response.ok) {
                        this.services = await response.json();
                    }
                } else {
                    let response = await fetch('api/containers');
                    if (response.ok) {
                        this.containers = await response.json();
                    }
                }
            },
            action: async function (link) {
                let response = await fetch(link.href, { method: link.type });
                if (response.ok) {
                    switch (response.status) {
                        case 200:
                            this.log = await response.text()
                            break;
                        default:
                            break;
                    }
                    this.loadData();
                    return;
                }
            }
            // downloadLog: async function (link) {
            //     response = await fetch(link.href, { method: link.type });
            //     if (response.ok) {
            //         this.log = await response.text()
            //         return;
            //     }
            // }
        }
    });
}

window.addEventListener('load', init);

$(document).on('click', 'div.conman-settings .dropdown-menu', function (e) {
    e.stopPropagation();
});