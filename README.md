## Cloudflare Updater
Update a Cloudflare record each time the external IP address changes.

### Features
* Runs in a Docker container
* Easy to configure
* Lightweight
* Built-in cronjob that runs every minute

### Usage
1. Clone this repo
    ```
    git clone https://github.com/AddilAfzal/cloudflare-updater
    ```
    
2. Build the image
    ```
    docker build . -t cloudflare-updater
    ```

3. Create a container
    
    ```
    docker create --name="cloudflare-updater" cloudflare-updater \
            --targetZone=<TARGET_ZONE> \
            --targetRecord=<TARGET_RECORD> \
            --email=<EMAIL> \
            --apiKey=<API_KEY> 
    ```
    Replace placeholder values.
    
    * `<TARGET_ZONE>` is the zone/domain in which the target record falls under. `example.com`
    * `<TARGET_RECORD>` is the record name or sub-domain belonging to the zone. `home.example.com`
    * `<EMAIL>` is the Cloudflare email address that owns the domain\zone that is being updated. 
    * `<API_KEY>` is the Cloudflare API key associated with the email address. [How do I get an API key?](https://support.cloudflare.com/hc/en-us/articles/200167836-Managing-API-Tokens-and-Keys)
4. Start the container
    ```
    docker start cloudflare-updater
    ```
  
### Notes
Only supports IPv4 for now.
