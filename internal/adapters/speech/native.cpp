//go:build speech

// Captures bounded mono float PCM with miniaudio and decodes with the pinned whisper.cpp.
// Device callbacks only copy samples; the Go session owns each capture and decoder lifetime.
#define MA_NO_DECODING
#define MA_NO_ENCODING
#define MA_NO_RESOURCE_MANAGER
#define MA_NO_NODE_GRAPH
#define MA_NO_ENGINE
#define MINIAUDIO_IMPLEMENTATION
#include "../../../third_party/speech/whisper.cpp/examples/miniaudio.h"
#include "whisper.h"
#include <algorithm>
#include <atomic>
#include <cmath>
#include <cstdlib>
#include <cstring>
#include <mutex>
#include <string>
#include <vector>

struct Capture {
    ma_device device{};
    std::mutex mutex;
    std::vector<float> pcm;
};
constexpr size_t max_samples = 30 * WHISPER_SAMPLE_RATE;

static void record(ma_device *device, void *, const void *input, ma_uint32 frames) {
    auto *c = static_cast<Capture *>(device->pUserData);
    std::lock_guard<std::mutex> guard(c->mutex);
    size_t count = std::min<size_t>(frames, max_samples - c->pcm.size());
    auto *first = static_cast<const float *>(input);
    if (first && count) {
        c->pcm.insert(c->pcm.end(), first, first + count);
    }
}

extern "C" void *speech_capture() {
    auto *c = new Capture;
    c->pcm.reserve(max_samples);
    auto config = ma_device_config_init(ma_device_type_capture);
    config.capture.format = ma_format_f32;
    config.capture.channels = 1;
    config.sampleRate = WHISPER_SAMPLE_RATE;
    config.dataCallback = record;
    config.pUserData = c;
    if (ma_device_init(nullptr, &config, &c->device) != MA_SUCCESS) {
        delete c;
        return nullptr;
    }
    if (ma_device_start(&c->device) != MA_SUCCESS) {
        ma_device_uninit(&c->device);
        delete c;
        return nullptr;
    }
    return c;
}

extern "C" void speech_stop(Capture *c) {
    ma_device_stop(&c->device);
}
extern "C" void speech_close_capture(Capture *c) {
    ma_device_uninit(&c->device);
    delete c;
}
extern "C" int speech_snapshot(Capture *c, float *out) {
    std::lock_guard<std::mutex> guard(c->mutex);
    if (out) {
        std::copy(c->pcm.begin(), c->pcm.end(), out);
    }
    return static_cast<int>(c->pcm.size());
}

static void quiet(enum ggml_log_level, const char *, void *) {}
extern "C" void *speech_load(const char *file) {
    whisper_log_set(quiet, nullptr);
    auto config = whisper_context_default_params();
    config.use_gpu = false;
    return whisper_init_from_file_with_params(file, config);
}
extern "C" void speech_free(whisper_context *ctx) {
    whisper_free(ctx);
}
extern "C" char *speech_decode(whisper_context *ctx, const float *audio, int count) {
    double energy = 0;
    for (int i = 0; i < count; ++i) {
        energy += audio[i] * audio[i];
    }
    if (count < WHISPER_SAMPLE_RATE / 4 || energy < count * 0.00000025) {
        return strdup("");
    }
    auto p = whisper_full_default_params(WHISPER_SAMPLING_GREEDY);
    p.n_threads = 4;
    p.no_context = true;
    p.no_timestamps = true;
    p.print_progress = false;
    p.print_realtime = false;
    p.print_timestamps = false;
    p.suppress_nst = true;
    p.language = whisper_is_multilingual(ctx) ? "auto" : "en";
    if (whisper_full(ctx,p,audio,count) != 0) {
        return nullptr;
    }
    std::string text;
    int segments = whisper_full_n_segments(ctx);
    for (int i = 0; i < segments; ++i) {
        text += whisper_full_get_segment_text(ctx,i);
    }
    return strdup(text.c_str());
}
