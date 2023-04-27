
1. Omalla koneella:
Grafana dashboard päällä (mulla brew avulla asennettu https://grafana.com/docs/grafana/latest/setup-grafana/installation/mac/). Tässä Linuxille: https://grafana.com/docs/grafana/latest/setup-grafana/installation/debian/. Tuon voi toki asentaa dockerin/kuberneteksen kautta myös, mutta menee helpommin tälleen, koska Python API runnaa localhost:8000 ja sinne pitäisi pystyä laittamaan requesteja.

Tämän jälkeen tarvitset omalle koneelle tämän json filun: https://github.com/luukasmakila/cri-resource-manager/blob/memtier-nri/demo/memtierd/memtierd-demo-grafana-dashboard.json
Jonka voit grafanaan importata suoraan Dashboards osiosta ja sieltä "New" ja "Import"

2. VM:ssä:
Pitäisi tehdä tuo port-forwardaus jonka ohjeet annoit aiemmin. Sitten laittaa python api runnaamaan:
cd prometheus_backend
source .venv/bin/activate
uvicorn main:app --reload

Sitten toiseen terminaaliin:
sudo /opt/nri/plugins/10-memtier-nri

(Siinä podi deploymentissa on jotain ongelmia sen kanssa että saan memtierd tulokset printtaamaan /tmp kansioon hostilla niin demoon tällä hetkellä kannattaa vain runnata toi binääri)

3. Nyt pitäisi olla setti kasassa. Jos restarttaat demon myöhemmmin niin prometheus_backend/data kansiosta pitäisi tyhjentää ne aiemmat entryt käsin tällä hetkellä.
Eli time series filut pitäisi näyttää tältä kun ovat tyhjiä:
Lowprio
{ 
    "time_series_lowprio_1": [ 
    ] 
}

Highprio
{ 
    "time_series_highprio_1": [ 
    ] 
}

Ja page fault filut tältä:
Highprio
{
    "page_faults_highprio_1": [
    ]
}

Lowprio:
{
    "page_faults_lowprio_1": [
    ]
}

Python koodi olettaa että siellä on tuossa muodossa json filut jos ovat tyhjinä eli esim tälläinen json filu hajottaisi sen:
{}

4. Sitten voi alkaa Grafanasta lähettää requesteja http://localhost:8000/metrics jonne tuo json modeli dashboardista jo valmiiksi osoittaa niin ei tarvitse muutoksia tehdä