package beater

import (
    "errors"
    "fmt"
    "net/url"
    "net"
    "strings"
    "time"

    "github.com/elastic/beats/libbeat/beat"
    "github.com/elastic/beats/libbeat/common"
    "github.com/elastic/beats/libbeat/logp"
    "github.com/elastic/beats/libbeat/publisher"

    "github.com/consulthys/logstashbeat/config"
)

const selector = "logstashbeat"

type Logstashbeat struct {
    done            chan struct{}
    client          publisher.Client
    config          config.Config

    urls            []*url.URL

    hotThreads      int

    jvmStats        bool
    processStats    bool
    pipelineStats   bool
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
    config := config.DefaultConfig
    if err := cfg.Unpack(&config); err != nil {
        return nil, fmt.Errorf("Error reading config file: %v", err)
    }

    bt := &Logstashbeat{
        done: make(chan struct{}),
        config: config,
    }

    //define default URL if none provided
    var urlConfig []string
    if config.URLs != nil {
        urlConfig = config.URLs
    } else {
        urlConfig = []string{"http://127.0.0.1:9600"}
    }

    bt.urls = make([]*url.URL, len(urlConfig))
    for i := 0; i < len(urlConfig); i++ {
        u, err := url.Parse(urlConfig[i])
        if err != nil {
            logp.Err("Invalid Logstash URL: %v", err)
            return nil, err
        }
        bt.urls[i] = u
    }

    bt.hotThreads = config.Hot_threads

    if config.Stats.JVM != nil {
        bt.jvmStats = *config.Stats.JVM
    } else {
        bt.jvmStats = true
    }

    if config.Stats.Process != nil {
        bt.processStats = *config.Stats.Process
    } else {
        bt.processStats = true
    }

    if config.Stats.Pipeline != nil {
        bt.pipelineStats = *config.Stats.Pipeline
    } else {
        bt.pipelineStats = true
    }

    if bt.hotThreads == 0  && !bt.jvmStats && !bt.processStats && !bt.pipelineStats {
        return nil, errors.New("Invalid statistics configuration")
    }

    logp.Info("Init logstashbeat")
    logp.Info("Period %v\n", bt.config.Period)
    logp.Info("Watch %v", bt.urls)
    logp.Info("Capture %v hot threads\n", bt.hotThreads)
    logp.Info("JVM statistics %t\n", bt.jvmStats)
    logp.Info("Process statistics %t\n", bt.processStats)
    logp.Info("Pipeline statistics %t\n", bt.pipelineStats)

    return bt, nil
}

func (bt *Logstashbeat) Run(b *beat.Beat) error {
    logp.Info("logstashbeat is running! Hit CTRL-C to stop it.")

    bt.client = b.Publisher.Connect()

    for _, u := range bt.urls {
        go func(u *url.URL) {

            s := strings.Split(u.Host, ":")
            host, port := s[0], s[1]

            ticker := time.NewTicker(bt.config.Period)
            counter := 1
            for {
                select {
                case <-bt.done:
                    goto GotoFinish
                case <-ticker.C:
                }

                timerStart := time.Now()

                // We want iterate for each dns resolved IP
                ipAddresses, _ := net.LookupHost(host)
                for i := 0; i < len(ipAddresses); i++  {
                    ipAddress := strings.SplitAfter(ipAddresses[i], " ")
                    u.Host = ipAddress[0] + ":" + port

                    if bt.hotThreads > 0 {
                        logp.Debug(selector, "Hot threads for url: %v", u)
                        hot_threads, err := bt.GetHotThreads(*u, bt.hotThreads)

                        if err != nil {
                            logp.Err("Error retrieving hot threads: %v", err)
                        } else {
                            logp.Debug(selector, "Hot threads detail: %+v", hot_threads)

                            event := common.MapStr{
                                "@timestamp":   common.Time(time.Now()),
                                "type":         "hot_threads",
                                "counter":      counter,
                                "ipv4":         u.Host,
                                "domain":       host,
                                "hot_threads":  hot_threads,
                            }

                            bt.client.PublishEvent(event)
                            logp.Debug(selector, "Logstash hot threads sent")
                            counter++
                        }
                    }

                    if bt.jvmStats {
                        logp.Debug(selector, "JVM stats for url: %v", u)
                        jvm_stats, err := bt.GetJvmStats(*u)

                        if err != nil {
                            logp.Err("Error reading JVM stats: %v", err)
                        } else {
                            logp.Debug(selector, "JVM stats detail: %+v", jvm_stats)

                            event := common.MapStr{
                                "@timestamp":   common.Time(time.Now()),
                                "type":         "jvm",
                                "counter":      counter,
                                "ipv4":         u.Host,
                                "domain":       host,
                                "jvm":          jvm_stats.Jvm,
                            }

                            bt.client.PublishEvent(event)
                            logp.Debug(selector, "Logstash JVM stats sent")
                            counter++
                        }
                    }

                    if bt.processStats {
                        logp.Debug(selector, "Process stats for url: %v", u)
                        process_stats, err := bt.GetProcessStats(*u)

                        if err != nil {
                            logp.Err("Error reading process stats: %v", err)
                        } else {
                            logp.Debug(selector, "Process stats detail: %+v", process_stats)

                            event := common.MapStr{
                                "@timestamp": common.Time(time.Now()),
                                "type":       "process",
                                "counter":    counter,
                                "ipv4":       u.Host,
                                "domain":     host,
                                "process":    process_stats.Process,
                            }

                            bt.client.PublishEvent(event)
                            logp.Debug(selector, "Logstash process stats sent")
                            counter++
                        }
                    }

                    if bt.pipelineStats {
                        logp.Debug(selector, "Pipeline stats for url: %v", u)
                        pipeline_stats, err := bt.GetPipelineStats(*u)

                        if err != nil {
                            logp.Err("Error reading pipeline stats: %v", err)
                        } else {
                            logp.Debug(selector, "Pipeline stats detail: %+v", pipeline_stats)

                            event := common.MapStr{
                                "@timestamp": common.Time(time.Now()),
                                "type":       "pipeline",
                                "counter":    counter,
                                "ipv4":       u.Host,
                                "domain":     host,
                                "pipeline":   pipeline_stats.Pipeline,
                            }

                            bt.client.PublishEvent(event)
                            logp.Debug(selector, "Logstash pipeline stats sent")
                            counter++
                        }
                    }
                }
                timerEnd := time.Now()
                duration := timerEnd.Sub(timerStart)
                if duration.Nanoseconds() > bt.config.Period.Nanoseconds() {
                    logp.Warn("Ignoring tick(s) due to processing taking longer than one period")
                }
            }
        GotoFinish:
        }(u)
    }

    <-bt.done
    return nil
}

func (bt *Logstashbeat) Stop() {
    logp.Info(selector, "Stop logstashbeat")
    bt.client.Close()
    close(bt.done)
}
