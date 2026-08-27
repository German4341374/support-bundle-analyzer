resource "docker_network" "analyzer" {
  name = "support-bundle-analyzer"
}

resource "docker_volume" "workspaces" {
  name = "support-bundle-analyzer-workspaces"
}

resource "docker_image" "analyzer" {
  name         = var.image
  keep_locally = true
}

resource "docker_container" "analyzer" {
  name  = "support-bundle-analyzer"
  image = docker_image.analyzer.image_id

  env = [
    "SBA_HOST=0.0.0.0",
    "SBA_ALLOW_REMOTE=true",
    "SBA_ACCESS_TOKEN=${var.access_token}",
  ]

  ports {
    internal = 8080
    external = var.host_port
    ip       = "127.0.0.1"
  }

  volumes {
    volume_name    = docker_volume.workspaces.name
    container_path = "/data/workspaces"
  }

  networks_advanced {
    name = docker_network.analyzer.name
  }

  read_only = true

  healthcheck {
    test     = ["CMD", "node", "-e", "fetch('http://127.0.0.1:8080/health').then(r=>{if(!r.ok)process.exit(1)}).catch(()=>process.exit(1))"]
    interval = "10s"
    timeout  = "3s"
    retries  = 5
  }
}
