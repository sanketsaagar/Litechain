# LightBeam Testnet Kurtosis Configuration
# This file defines a multi-node testing network for LightBeam (LightChain L1 Testnet)

# Constants
LIGHTCHAIN_IMAGE = "lightchain:latest"
POSTGRES_IMAGE = "postgres:15-alpine"
GRAFANA_IMAGE = "grafana/grafana:latest"
PROMETHEUS_IMAGE = "prom/prometheus:latest"

# Network configuration
NETWORK_ID = 1337
CHAIN_ID = 1337

# Node configurations
VALIDATOR_NODES = 3
FULLNODE_NODES = 2
ARCHIVE_NODES = 1

# Continuous operation features
ENABLE_AUTO_MINING = True
ENABLE_TX_GENERATION = True
ENABLE_CONTINUOUS_OPERATION = True
TX_GENERATOR_IMAGE = "python:3.11-alpine"

def run(plan, args={}):
    """
    Main function to deploy LightBeam testnet
    """
    
    plan.print("Starting LightBeam testnet deployment...")
    
    # Parse arguments
    validator_count = args.get("validators", VALIDATOR_NODES)
    fullnode_count = args.get("fullnodes", FULLNODE_NODES)  
    archive_count = args.get("archives", ARCHIVE_NODES)
    enable_monitoring = args.get("monitoring", True)
    
    # Deploy infrastructure services
    postgres_service = deploy_postgres(plan)
    
    # Deploy monitoring stack
    if enable_monitoring:
        monitoring_services = deploy_monitoring(plan)
    
    # Generate genesis configuration
    genesis_config = generate_genesis_config(validator_count)
    
    # Deploy validator nodes
    validator_services = []
    for i in range(validator_count):
        validator = deploy_validator_node(plan, i, genesis_config, postgres_service)
        validator_services.append(validator)
    
    # Deploy sequencer nodes  
    sequencer_services = []
    for i in range(sequencer_count):
        sequencer = deploy_sequencer_node(plan, i, genesis_config, postgres_service)
        sequencer_services.append(sequencer)
        
    # Deploy archive nodes
    archive_services = []
    for i in range(archive_count):
        archive = deploy_archive_node(plan, i, genesis_config, postgres_service)
        archive_services.append(archive)
    
    # Deploy load balancer for RPC endpoints
    load_balancer = deploy_load_balancer(plan, validator_services + sequencer_services)
    
    # Setup network connections
    setup_network_topology(plan, validator_services, sequencer_services, archive_services)
    
    # Deploy continuous operation services
    if args.get("enable_tx_generation", ENABLE_TX_GENERATION):
        tx_generator = deploy_transaction_generator(plan, load_balancer)
    
    # Deploy network lifecycle manager
    lifecycle_manager = deploy_lifecycle_manager(plan, validator_services, sequencer_services, archive_services)
    
    # Deploy test tools
    deploy_test_tools(plan, load_balancer)
    
    plan.print("🎉 LightChain L2 devnet deployment completed!")
    plan.print("🌐 Access Points:")
    plan.print("   • RPC endpoint: http://localhost:8545")
    plan.print("   • WebSocket endpoint: ws://localhost:8546")
    if enable_monitoring:
        plan.print("   • Grafana dashboard: http://localhost:3000 (admin/admin123)")
        plan.print("   • Prometheus metrics: http://localhost:9090")
    plan.print("🔥 Continuous Features:")
    plan.print("   • Auto-mining: ENABLED (blocks every 1-2s)")
    plan.print("   • Transaction generation: ENABLED (realistic patterns)")
    plan.print("   • Network persistence: ENABLED (until upgrade)")
    plan.print("🎮 Management:")
    plan.print("   • Status: kurtosis service logs lightchain-devnet lifecycle-manager")
    plan.print("   • Upgrade: kurtosis service exec lightchain-devnet lifecycle-manager trigger_upgrade")
    plan.print("   • Monitor: kurtosis port print lightchain-devnet grafana 3000")

def deploy_postgres(plan):
    """Deploy PostgreSQL database for shared storage"""
    
    postgres_config = ServiceConfig(
        image=POSTGRES_IMAGE,
        ports={
            "postgres": PortSpec(5432, application_protocol="postgresql"),
        },
        env_vars={
            "POSTGRES_DB": "lightchain",
            "POSTGRES_USER": "lightchain",
            "POSTGRES_PASSWORD": "lightchain123",
        },
        files={
            "/docker-entrypoint-initdb.d": Directory(
                artifact_names=[render_templates(
                    config_template="postgres-init.sql.tmpl",
                    name="postgres-init-sql",
                    template_data={}
                )]
            )
        }
    )
    
    postgres_service = plan.add_service(
        name="postgres",
        config=postgres_config
    )
    
    # Wait for database to be ready
    plan.wait(
        service_name="postgres",
        recipe=ExecRecipe(command=["pg_isready", "-U", "lightchain"]),
        field="code",
        assertion="==",
        target_value=0,
        timeout="30s"
    )
    
    return postgres_service

def deploy_monitoring(plan):
    """Deploy monitoring stack with Prometheus and Grafana"""
    
    # Deploy Prometheus
    prometheus_config = ServiceConfig(
        image=PROMETHEUS_IMAGE,
        ports={
            "prometheus": PortSpec(9090, application_protocol="http"),
        },
        files={
            "/etc/prometheus": Directory(
                artifact_names=[render_templates(
                    config_template="prometheus.yml.tmpl",
                    name="prometheus-config", 
                    template_data={}
                )]
            )
        }
    )
    
    prometheus_service = plan.add_service(
        name="prometheus",
        config=prometheus_config
    )
    
    # Deploy Grafana
    grafana_config = ServiceConfig(
        image=GRAFANA_IMAGE,
        ports={
            "grafana": PortSpec(3000, application_protocol="http"),
        },
        env_vars={
            "GF_SECURITY_ADMIN_PASSWORD": "admin123",
        },
        files={
            "/etc/grafana/provisioning": Directory(
                artifact_names=[render_templates(
                    config_template="grafana-provisioning.tmpl",
                    name="grafana-provisioning",
                    template_data={"prometheus_url": "http://prometheus:9090"}
                )]
            )
        }
    )
    
    grafana_service = plan.add_service(
        name="grafana", 
        config=grafana_config
    )
    
    return {
        "prometheus": prometheus_service,
        "grafana": grafana_service
    }

def deploy_validator_node(plan, node_id, genesis_config, postgres_service):
    """Deploy a validator node"""
    
    node_name = "validator-{}".format(node_id)
    
    # Generate node configuration with continuous operation features
    node_config = render_templates(
        config_template="validator-node.yaml.tmpl",
        name="{}-config".format(node_name),
        template_data={
            "node_id": node_id,
            "postgres_host": postgres_service.hostname,
            "genesis_config": genesis_config,
            "rpc_port": 8545 + node_id,
            "ws_port": 8546 + node_id,
            "p2p_port": 30303 + node_id,
            "metrics_port": 9090 + node_id,
            "enable_auto_mining": ENABLE_AUTO_MINING,
            "auto_mine_blocks": True,
            "dev_period": "2s",
            "generate_empty_blocks": True,
            "continuous_mining": True
        }
    )
    
    service_config = ServiceConfig(
        image=LIGHTCHAIN_IMAGE,
        ports={
            "rpc": PortSpec(8545 + node_id, application_protocol="http"),
            "ws": PortSpec(8546 + node_id, application_protocol="ws"),
            "p2p": PortSpec(30303 + node_id, application_protocol="tcp"),
            "metrics": PortSpec(9090 + node_id, application_protocol="http")
        },
        files={
            "/app/config": Directory(artifact_names=[node_config]),
            "/app/genesis": Directory(artifact_names=[genesis_config])
        },
        cmd=[
            "/app/lightchain",
            "--config", "/app/config/validator.yaml",
            "--type", "validator"
        ]
    )
    
    validator_service = plan.add_service(
        name=node_name,
        config=service_config
    )
    
    # Wait for node to be ready
    plan.wait(
        service_name=node_name,
        recipe=HttpRequestRecipe(
            port_id="rpc",
            method="POST", 
            endpoint="/",
            content_type="application/json",
            body='{"jsonrpc":"2.0","method":"net_version","params":[],"id":1}'
        ),
        field="code",
        assertion="==", 
        target_value=200,
        timeout="60s"
    )
    
    return validator_service

def deploy_sequencer_node(plan, node_id, genesis_config, postgres_service):
    """Deploy a sequencer node"""
    
    node_name = "sequencer-{}".format(node_id)
    
    # Generate node configuration with continuous operation features
    node_config = render_templates(
        config_template="sequencer-node.yaml.tmpl", 
        name="{}-config".format(node_name),
        template_data={
            "node_id": node_id,
            "postgres_host": postgres_service.hostname,
            "genesis_config": genesis_config,
            "rpc_port": 8555 + node_id,
            "ws_port": 8556 + node_id, 
            "p2p_port": 30313 + node_id,
            "metrics_port": 9091 + node_id,
            "enable_auto_mining": ENABLE_AUTO_MINING,
            "auto_mine_blocks": True,
            "dev_period": "1s",
            "generate_empty_blocks": True,
            "continuous_mining": True,
            "background_mining": True
        }
    )
    
    service_config = ServiceConfig(
        image=LIGHTCHAIN_IMAGE,
        ports={
            "rpc": PortSpec(8555 + node_id, application_protocol="http"),
            "ws": PortSpec(8556 + node_id, application_protocol="ws"),
            "p2p": PortSpec(30313 + node_id, application_protocol="tcp"),
            "metrics": PortSpec(9091 + node_id, application_protocol="http")
        },
        files={
            "/app/config": Directory(artifact_names=[node_config]),
            "/app/genesis": Directory(artifact_names=[genesis_config])
        },
        cmd=[
            "/app/lightchain",
            "--config", "/app/config/sequencer.yaml", 
            "--type", "sequencer"
        ]
    )
    
    sequencer_service = plan.add_service(
        name=node_name,
        config=service_config
    )
    
    return sequencer_service

def deploy_archive_node(plan, node_id, genesis_config, postgres_service):
    """Deploy an archive node"""
    
    node_name = "archive-{}".format(node_id)
    
    # Generate node configuration
    node_config = render_templates(
        config_template="archive-node.yaml.tmpl",
        name="{}-config".format(node_name),
        template_data={
            "node_id": node_id,
            "postgres_host": postgres_service.hostname,
            "genesis_config": genesis_config, 
            "rpc_port": 8565 + node_id,
            "ws_port": 8566 + node_id,
            "p2p_port": 30323 + node_id,
            "metrics_port": 9092 + node_id
        }
    )
    
    service_config = ServiceConfig(
        image=LIGHTCHAIN_IMAGE,
        ports={
            "rpc": PortSpec(8565 + node_id, application_protocol="http"),
            "ws": PortSpec(8566 + node_id, application_protocol="ws"), 
            "p2p": PortSpec(30323 + node_id, application_protocol="tcp"),
            "metrics": PortSpec(9092 + node_id, application_protocol="http")
        },
        files={
            "/app/config": Directory(artifact_names=[node_config]),
            "/app/genesis": Directory(artifact_names=[genesis_config])
        },
        cmd=[
            "/app/lightchain",
            "--config", "/app/config/archive.yaml",
            "--type", "archive"
        ]
    )
    
    archive_service = plan.add_service(
        name=node_name,
        config=service_config
    )
    
    return archive_service

def deploy_load_balancer(plan, node_services):
    """Deploy load balancer for RPC endpoints"""
    
    # Create nginx configuration for load balancing
    upstream_servers = []
    for service in node_services:
        upstream_servers.append("{}:8545".format(service.hostname))
    
    nginx_config = render_templates(
        config_template="nginx.conf.tmpl",
        name="nginx-config",
        template_data={
            "upstream_servers": upstream_servers
        }
    )
    
    load_balancer_config = ServiceConfig(
        image="nginx:alpine",
        ports={
            "http": PortSpec(8545, application_protocol="http"),
            "ws": PortSpec(8546, application_protocol="ws")
        },
        files={
            "/etc/nginx": Directory(artifact_names=[nginx_config])
        }
    )
    
    return plan.add_service(
        name="load-balancer",
        config=load_balancer_config
    )

def setup_network_topology(plan, validators, sequencers, archives):
    """Setup P2P network connections between nodes"""
    
    # TODO: Implement peer discovery and connection setup
    # This would involve configuring bootstrap nodes and peer connections
    pass

def deploy_test_tools(plan, load_balancer):
    """Deploy testing and development tools"""
    
    # Deploy a simple web interface for testing
    test_tools_config = ServiceConfig(
        image="node:18-alpine",
        ports={
            "web": PortSpec(3001, application_protocol="http")
        },
        cmd=[
            "sh", "-c", 
            "npx create-react-app test-app && cd test-app && npm start"
        ]
    )
    
    plan.add_service(
        name="test-tools",
        config=test_tools_config
    )

def generate_genesis_config(validator_count):
    """Generate genesis configuration"""
    
    # TODO: Generate proper genesis block with validator set
    genesis_data = {
        "chain_id": CHAIN_ID,
        "network_id": NETWORK_ID,
        "validators": [],
        "genesis_time": "2024-01-01T00:00:00Z",
        "initial_supply": "1000000000000000000000000000"
    }
    
    return render_templates(
        config_template="genesis.json.tmpl", 
        name="genesis-config",
        template_data=genesis_data
    )

def deploy_transaction_generator(plan, load_balancer_service):
    """Deploy continuous transaction generator for realistic blockchain activity"""
    
    # Create transaction generator script as an artifact
    tx_generator_script = plan.render_templates(
        config={
            "continuous_tx_generator.py": struct(
                template=read_file("/scripts/continuous-tx-generator.py"),
                data={}
            )
        },
        name="tx-generator-script"
    )
    
    tx_generator_config = ServiceConfig(
        image=TX_GENERATOR_IMAGE,
        files={
            "/app/scripts": Directory(artifact_names=[tx_generator_script])
        },
        env_vars={
            "PYTHONUNBUFFERED": "1",
            "RPC_URL": "http://{}:8545".format(load_balancer_service.hostname),
            "TX_INTERVAL": "3",
            "ENABLE_BURST_MODE": "true"
        },
        cmd=[
            "sh", "-c",
            "pip install requests && python /app/scripts/continuous_tx_generator.py"
        ]
    )
    
    tx_generator_service = plan.add_service(
        name="tx-generator",
        config=tx_generator_config
    )
    
    plan.print("🚀 Transaction generator deployed - generating realistic blockchain activity")
    return tx_generator_service

def deploy_lifecycle_manager(plan, validators, sequencers, archives):
    """Deploy network lifecycle manager for continuous operation"""
    
    # Create lifecycle management script
    lifecycle_script = plan.render_templates(
        config={
            "lifecycle_manager.py": struct(
                template="""
import time
import requests
import json
import os
from datetime import datetime

class LifecycleManager:
    def __init__(self):
        self.running = True
        self.upgrade_required = False
        
    def monitor_network(self):
        while self.running:
            try:
                # Check if upgrade flag exists
                if os.path.exists('/app/NETWORK_UPGRADE_REQUIRED'):
                    self.trigger_graceful_upgrade()
                    break
                    
                # Monitor network health
                self.check_network_health()
                time.sleep(60)  # Check every minute
                
            except Exception as e:
                print(f"Monitor error: {e}")
                time.sleep(60)
    
    def check_network_health(self):
        # Check RPC endpoints
        healthy_nodes = 0
        total_nodes = len([""" + str(len(validators + sequencers)) + """])
        
        for i, node in enumerate([""" + ",".join(['"{}"'.format(v.hostname) for v in validators + sequencers]) + """]):
            try:
                response = requests.post(
                    f"http://{node}:8545",
                    json={"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1},
                    timeout=5
                )
                if response.status_code == 200:
                    healthy_nodes += 1
            except:
                pass
        
        health_ratio = healthy_nodes / total_nodes if total_nodes > 0 else 0
        timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        
        print(f"[{timestamp}] Network Health: {healthy_nodes}/{total_nodes} nodes healthy ({health_ratio:.1%})")
        
        if health_ratio < 0.5:
            print(f"[{timestamp}] WARNING: Network degraded - less than 50% nodes healthy")
    
    def trigger_graceful_upgrade(self):
        print("[UPGRADE] Network upgrade requested - initiating graceful shutdown")
        # In a real implementation, this would coordinate with all nodes
        self.running = False
        
    def trigger_upgrade(self):
        with open('/app/NETWORK_UPGRADE_REQUIRED', 'w') as f:
            f.write(f"Upgrade requested at {datetime.now()}")
        print("Upgrade flag created - network will shutdown gracefully")

if __name__ == "__main__":
    manager = LifecycleManager()
    
    import sys
    if len(sys.argv) > 1 and sys.argv[1] == "trigger_upgrade":
        manager.trigger_upgrade()
    else:
        print("🔍 Starting LightChain L2 Lifecycle Manager")
        print("   Monitoring network health and managing upgrades")
        manager.monitor_network()
""",
                data={}
            )
        },
        name="lifecycle-manager-script"
    )
    
    lifecycle_config = ServiceConfig(
        image="python:3.11-alpine",
        files={
            "/app": Directory(artifact_names=[lifecycle_script])
        },
        env_vars={
            "PYTHONUNBUFFERED": "1"
        },
        cmd=[
            "sh", "-c",
            "pip install requests && python /app/lifecycle_manager.py"
        ]
    )
    
    lifecycle_service = plan.add_service(
        name="lifecycle-manager",
        config=lifecycle_config
    )
    
    plan.print("🛡️ Lifecycle manager deployed - monitoring network health")
    return lifecycle_service

def render_templates(config_template, name, template_data):
    """Helper function to render configuration templates"""
    
    # TODO: Implement actual template rendering
    # For now, return a placeholder
    return name
