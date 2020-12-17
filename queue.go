package main

// QueueGetSong
func (v *VoiceInstance) QueueGetSpeech() (speech Speech) {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()
	if len(v.queue) != 0 {
		return v.queue[0]
	}
	return
}

// QueueAdd
func (v *VoiceInstance) QueueAdd(speech Speech) {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()
	v.queue = append(v.queue, speech)
}

// QueueClean
func (v *VoiceInstance) QueueClean() {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()
	v.queue = []Speech{}
}

// QueueRemoveFirst
func (v *VoiceInstance) QueueRemoveFisrt() {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()
	if len(v.queue) != 0 {
		v.queue = v.queue[1:]
	}
}
