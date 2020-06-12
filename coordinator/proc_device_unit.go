package main

import (
    "fmt"
    "strconv"
    "strings"
    log "github.com/sirupsen/logrus"
)

func proc_device_ios_unit( o ProcOptions, uuid string, curIP string) {
    vncPort := 0
    if o.config.Video.UseVnc && o.config.Video.Enabled {
        vncPort = o.devd.vncPort
    }
    
    secure := o.config.FrameServer.Secure
    var frameServer string
    var storageUrl string
    if secure {
        frameServer = fmt.Sprintf("wss://%s:%d/echo", curIP, o.devd.vidPort)
        storageUrl = "https://%s"
    } else {
        frameServer = fmt.Sprintf("ws://%s:%d/echo", curIP, o.devd.vidPort)
        storageUrl = "http://%s"
    }
    
    o.args = []string{
        fmt.Sprintf("--inspect=0.0.0.0:%d", o.devd.devIosPort),
        "runmod.js"              , "device-ios",
        "--serial"               , uuid,
        "--name"                 , o.devd.name,
        "--connect-push"         , fmt.Sprintf("tcp://%s:7270", o.config.Stf.Ip),
        "--connect-sub"          , fmt.Sprintf("tcp://%s:7250", o.config.Stf.Ip),
        "--public-ip"            , curIP,
        "--wda-port"             , strconv.Itoa( o.devd.wdaPort ),
        "--storage-url"          , fmt.Sprintf(storageUrl, serverHostname),
        "--screen-ws-url-pattern", fmt.Sprintf("wss://%s/frames/%s/%d/x", o.config.Stf.HostName, curIP, o.devd.vidPort),
        //"--screen-ws-url-pattern", frameServer,
        "--vnc-password"         , o.config.Video.VncPassword,
        "--vnc-port"             , strconv.Itoa( vncPort ),
        "--vnc-scale"            , strconv.Itoa( o.config.Video.VncScale ),
        "--stream-width"         , strconv.Itoa( o.devd.streamWidth ),
        "--stream-height"        , strconv.Itoa( o.devd.streamHeight ),
        "--click-width"          , strconv.Itoa( o.devd.clickWidth ),
        "--click-height"         , strconv.Itoa( o.devd.clickHeight ),
        "--click-scale"          , strconv.Itoa( o.devd.clickScale ),
    }
    o.startFields = log.Fields {
        "server_ip": o.config.Stf.Ip,
        "client_ip": curIP,
        "server_host": o.config.Stf.HostName,
        "video_port": o.devd.vidPort,
        "node_port": o.devd.devIosPort,
        "device_name": o.devd.name,
        "vnc_scale": o.config.Video.VncScale,
        "stream_width": o.devd.streamWidth,
        "stream_height": o.devd.streamHeight,
        "clickScale": o.devd.clickScale,
        "clickWidth": o.devd.clickWidth,
        "clickHeight": o.devd.clickHeight,
        "frame_server": frameServer,
    }
    o.stdoutHandler = func( line string, plog *log.Entry ) (bool) {
        if strings.Contains( line, "Now owned by" ) {
            pos := strings.Index( line, "Now owned by" )
            pos += len( "Now owned by" ) + 2
            ownedStr := line[ pos: ]
            endpos := strings.Index( ownedStr, "\"" )
            owner := ownedStr[ :endpos ]
            plog.WithFields( log.Fields{
                "type": "wda_owner_start",
                "owner": owner,
            } ).Info("Device Owner Start")
        }
        if strings.Contains( line, "No longer owned by" ) {
            pos := strings.Index( line, "No longer owned by" )
            pos += len( "No longer owned by" ) + 2
            ownedStr := line[ pos: ]
            endpos := strings.Index( ownedStr, "\"" )
            owner := ownedStr[ :endpos ]
            plog.WithFields( log.Fields{
                "type": "wda_owner_stop",
                "owner": owner,
            } ).Info("Device Owner Stop")
        }
        if strings.Contains( line, "responding with identity" ) {
            plog.WithFields( log.Fields{
                "type": "device_ios_ident",
            } ).Debug("Device IOS Unit Registered Identity")
        }
        if strings.Contains( line, "Sent ready message" ) {
            plog.WithFields( log.Fields{
                "type": "device_ios_ready",
            } ).Debug("Device IOS Unit Ready")
        }
        return true
    }
    o.procName = "stf_device_ios"
    o.binary = "/usr/local/opt/node@12/bin/node"
    o.startDir = "./repos/stf-ios-provider"
    proc_generic( o )
}
            
