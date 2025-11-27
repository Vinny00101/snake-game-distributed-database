<h1 align="center"> Bem-vindo ao projeto em Go com Banco de dados distribuidos com Docker Swarm e MongoDB </h1>

<div align="center">
  <img src="https://img.shields.io/badge/MongoDB-47A248?style=for-the-badge&logo=mongodb&logoColor=white" alt="MongoDB Badge">
  &nbsp;&nbsp; 
  <img src="https://img.shields.io/badge/Docker%20Swarm-646272?style=for-for-the-badge&logo=docker&logoColor=white" alt="Docker Swarm Badge">
  &nbsp;&nbsp;
  <img src="https://img.shields.io/badge/Docker-2CA5E0?style=for-the-badge&logo=docker&logoColor=white" alt="Docker Swarm Badge">
  &nbsp;&nbsp;
  <img src="https://img.shields.io/badge/Docker%20Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker Swarm Badge">
  &nbsp;&nbsp;
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="GoLang Badge">
</div>

# Resumo

Projeto de Banco de Dados Distribuído com **Docker Swarm e MongoDB**
Este guia detalha a configuração de um **cluster Docker Swarm local** usando **Multipass** para simular um ambiente de produção com múltiplas **máquinas virtuais (VMs)**. O objetivo é rodar um **Replica Set de MongoDB**, demonstrando distribuição horizontal de um banco de dados **NoSQL**.


## 1 Pré-requisitos
- Docker Desktop (Com suporte a WSL 2/Hyper-V).

- Multipass (Instalado no Windows/macOS/Linux).

- Pasta do Projeto (Onde este README e o docker-compose.yml estão localizados).

## 2 Configuração Inicial do Cluster Swarm
Esta etapa cria as VMs e configura o Docker Swarm. Criar e Instalar o Docker nas VMs. Para economizar RAM, o Worker usa menos memória.

Crie as VMs (Manager e Workers):

```
Bash

multipass launch --name manager --cpus 2 --memory 2G --disk 15G
multipass launch --name worker1 --cpus 1 --memory 1G --disk 7G
multipass launch --name worker2 --cpus 1 --memory 1G --disk 7G

```

Para acessar cada shell das VMs deve ser utilizar do seguinte comando:


```
Bash

multipass shell <name-vm>

```

Após acessar o shell da VM deve ser fazer a instalacao do Docker, esse processo deve ser repertido em todas as VM, a seguir os comandos:


```
## dois comandos deve ser executado individualmente em cada VM.

multipass exec $vm -- bash -c "sudo apt-get update && sudo apt-get install -y docker.io"
multipass exec $vm -- bash -c "sudo usermod -aG docker ubuntu"
```

## 3 Inicializar o Swarm

Acesse o Manager:

```
Bash

multipass shell manager
```

Inicialize o Swarm (dentro da VM manager): Se o comando multipass info manager falhar dentro da VM, use o IP que aparece em multipass list:
```
Bash

# Se precisar usar o IP direto:
docker swarm init --advertise-addr <IP_DO_MANAGER>
```

Copie o Token: Copie o comando **docker swarm join...** impresso na tela. E saia do Manager.


```
Bash

exit
```

Junte os Workers (Fora das VMs):


```
Bash

multipass exec worker1 -- <COMANDO_JOIN_COMPLETO>
multipass exec worker2 -- <COMANDO_JOIN_COMPLETO>
```

## 3 Verificar e Compartilhar a Pasta

Verifique o Cluster (dentro da VM manager):

```
Bash

multipass shell manager
docker node ls
exit
```
Habilite e Monte a Pasta do Projeto: Execute no terminal principal para permitir que o Manager acesse o docker-compose.yml.
```
Bash

multipass set local.privileged-mounts=true
multipass mount C:\Caminho\Para\Sua\Pasta manager:/home/ubuntu/project
```

# Configuração de Rede no Host (Windows)

Este passo é obrigatório e deve ser executado no seu sistema Windows para permitir que sua API (que está rodando no Windows) consiga se conectar e resolver os nomes dos nós do MongoDB que estão no Cluster Docker Swarm.

Se este passo não for realizado, sua API receberá o erro: no such host (mongoX).

## 1. Mapeamento de Hostnames (DNS Local)
Precisamos mapear os nomes internos do serviço de banco de dados (mongo1, mongo2, mongo3) para os IPs das Máquinas Virtuais (VMs) correspondentes.

### 1.1 Obter os Endereços IP
Antes de editar, obtenha os endereços IPv4 atuais de todas as suas VMs. Execute no seu terminal (PowerShell/WSL):
Abra o Menu Iniciar e procure por "Bloco de Notas" (Notepad).

- Clique com o botão direito no ícone do Bloco de Notas e selecione "Executar como Administrador".

- No Bloco de Notas (como Admin), vá em Arquivo > Abrir.

- Navegue até o caminho: C:\Windows\System32\drivers\etc\

- Mude o filtro de arquivos de *.txt para "Todos os Arquivos (*.*)".

- Abra o arquivo chamado hosts.

```
Bash

# Mapeamento para o Cluster MongoDB (rs0) - Docker Swarm
# Permite que o Windows resolva os nomes internos da rede Overlay
<IP_WORKER1>   mongo1
<IP_WORKER2>   mongo2
<IP_MANAGER>   mongo3

```

Use o comando **multipass list** para pegar os IPs de:
- manager (onde está o mongo3);
- worker1 (onde está o mongo1);
- worker2 (onde está o mongo2).

# Deploy do MongoDB Replica Set
Faça o Deploy (dentro da VM manager):

```
Bash

multipass shell manager
cd /home/ubuntu/project
docker stack deploy -c docker-compose.yml meu_mongo
```
**Inicialize o Replica Set (Passo Manual):**
- Descubra o ID do contêiner mongo3 com **docker ps**.
- Acesse o shell do Mongo: **docker exec -it <ID_DO_MONGO3> mongosh**
- No shell do Mongo, execute:

```
JavaScript

rs.initiate({
   _id: "rs0",
   members: [
      { _id: 0, host: "mongo1:27017" },
      { _id: 1, host: "mongo2:27017" },
      { _id: 2, host: "mongo3:27017" }
   ]
})
```
O prompt deve mudar para rs0:PRIMARY>.

IV. Conexão da API
Sua API agora deve usar a seguinte Connection String para se conectar ao Replica Set (o driver fará o resto):

# Conexão da API
Sua API agora deve usar a seguinte Connection String para se conectar ao Replica Set (o driver fará o resto):

```
mongodb://mongo1:27017,mongo2:27017,mongo3:27017/?replicaSet=rs0
```
Se a configuração do hosts estiver correta, a API funcionará.