package v1

type CanaryPhase string

const (
    
    // Creation of managed resources has started
    CanaryPhaseInitializing CanaryPhase = "Initializing"
    
    // The canary analysis & traffic shifting has started
    CanaryPhaseProgressing CanaryPhase = "Progressing"
    
    // The canary analysis has finished,
    // and the canary deployment is being promoted
    CanaryPhasePromoting CanaryPhase = "Promoting"
    
    // The canary was promoted successfully
    CanaryPhaseSucceeded CanaryPhase = "Succeeded"
    
    // The canary failed and was rolled back
    CanaryPhaseFailed CanaryPhase = "Failed"
)
